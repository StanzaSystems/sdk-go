package fiberstanza

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/StanzaSystems/sdk-go/handlers/httphandler"
	"github.com/StanzaSystems/sdk-go/keys"
	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/StanzaSystems/sdk-go/stanza"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/trace"
)

// Client defines options for a new Stanza Client
type Client struct {
	// Required
	APIKey string // customer generated API key

	// Optional
	Name        string // defines applications name
	Release     string // defines applications version
	Environment string // defines applications environment
	StanzaHub   string // host:port (ipv4, ipv6, or resolvable hostname)

	Guard []string // prefetch config for these guards
}

// Optional arguments
type Opt struct {
	Headers       http.Header
	Feature       string
	PriorityBoost int32
	DefaultWeight float32
}

var (
	inboundHandler *httphandler.InboundHandler = nil
)

// New creates a new fiberstanza middleware fiber.Handler
func New(guardName string, opts ...Opt) fiber.Handler {
	if inboundHandler == nil {
		h, err := stanza.NewHttpInboundHandler()
		if err != nil {
			logging.Error(fmt.Errorf("failed to create HTTP inbound handler: %v", err))
			return func(c *fiber.Ctx) error {
				// with no InboundHandler there is nothing we can do but fail open
				logging.Error(fmt.Errorf("no HTTP inbound handler, failing open"))
				return c.Next()
			}
		}
		inboundHandler = h
	}
	h := inboundHandler

	return func(c *fiber.Ctx) error {
		var req http.Request
		if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
			logging.Error(fmt.Errorf("failed to convert request from fasthttp: %v", err))
			h.Meter().AllowedSuccessCount.Add(c.UserContext(), 1,
				[]metric.AddOption{metric.WithAttributes(append(h.Attributes(),
					h.ReasonFailOpen())...)}...)
			return c.Next() // log error and fail open
		}

		ctx := h.Propagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
		traceOpts := []trace.SpanStartOption{
			trace.WithAttributes(httpconv.ServerRequest("", &req)...),
			trace.WithSpanKind(trace.SpanKindServer),
		}
		ctx, span := h.Tracer().Start(ctx, guardName, traceOpts...)
		defer span.End()

		guard := h.NewGuard(addKeys(ctx, opts...), span, guardName, req.Header.Values("x-stanza-token"))
		c.SetUserContext(guard.Context())

		// Stanza Blocked
		if guard.Blocked() {
			span.SetStatus(codes.Error, guard.BlockMessage())
			c.SendString(guard.BlockMessage())
			return c.SendStatus(http.StatusTooManyRequests)
		}

		// Stanza Allowed
		err := c.Next() // intercept c.Next() for guard.End() status
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			guard.End(guard.Failure)
		} else {
			span.SetStatus(codes.Ok, "OK")
			guard.End(guard.Success)
		}
		return err
	}
}

// Init is a fiberstanza helper function (passthrough to stanza.Init)
func Init(ctx context.Context, client Client) (func(), error) {
	exit, err := stanza.Init(ctx, stanza.ClientOptions(client))
	if err != nil {
		return nil, err
	}
	return exit, nil
}

// HttpGet is a fiberstanza helper function (passthrough to stanza.HttpGet)
func HttpGet(req stanza.GuardRequest) (*http.Response, error) {
	return stanza.HttpGet(req)
}

// HttpPost is a fiberstanza helper function (passthrough to stanza.HttpPost)
func HttpPost(req stanza.GuardRequest, body io.Reader) (*http.Response, error) {
	req.Body = body
	return stanza.HttpPost(req)
}

// Guard is a fiberstanza helper function
func Guard(c *fiber.Ctx, name string, url string, opts ...Opt) stanza.GuardRequest {
	var req http.Request
	fasthttpadaptor.ConvertRequest(c.Context(), &req, true)
	ctx := otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
	guardRequest := stanza.GuardRequest{
		Context: ctx,
		Name:    name,
		URL:     url,
	}
	if len(opts) == 1 {
		guardRequest.Headers = opts[0].Headers
		guardRequest.Opt.Feature = opts[0].Feature
		guardRequest.Opt.PriorityBoost = opts[0].PriorityBoost
		guardRequest.Opt.DefaultWeight = opts[0].DefaultWeight
	}
	return guardRequest
}

func addKeys(ctx context.Context, opts ...Opt) context.Context {
	if len(opts) == 1 {
		if opts[0].Feature != "" {
			ctx = context.WithValue(ctx, keys.StanzaFeatureNameKey, opts[0].Feature)
		}
		if opts[0].PriorityBoost != 0 {
			ctx = context.WithValue(ctx, keys.StanzaPriorityBoostKey, opts[0].PriorityBoost)
		}
		if opts[0].DefaultWeight != 0 {
			ctx = context.WithValue(ctx, keys.StanzaDefaultWeightKey, opts[0].DefaultWeight)
		}
		if opts[0].Headers != nil {
			ctx = context.WithValue(ctx, keys.OutboundHeadersKey, opts[0].Headers)
		} else {
			ctx = context.WithValue(ctx, keys.OutboundHeadersKey, make(http.Header))
		}
	}
	return ctx
}
