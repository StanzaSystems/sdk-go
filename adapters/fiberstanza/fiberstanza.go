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

type guardRequest struct {
	ctx  context.Context
	name string
	url  string
	body io.Reader
}

var (
	inboundHandler  *httphandler.InboundHandler  = nil
	outboundHandler *httphandler.OutboundHandler = nil
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
			c.SendString("Stanza Inbound Rate Limited")
			return c.SendStatus(http.StatusTooManyRequests)
		}

		// Stanza Allowed
		err := c.Next() // intercept c.Next() for guard.End() status
		if err != nil {
			guard.End(guard.Failure)
		} else {
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
	if outboundHandler == nil {
		outboundHandler, err = stanza.NewHttpOutboundHandler()
		if err != nil {
			logging.Error(fmt.Errorf("failed to create HTTP outbound handler: %v", err))
			return nil, err
		}
	}
	return exit, nil
}

// HttpGet is a fiberstanza helper function (passthrough to stanza.NewHttpOutboundHandler)
func HttpGet(req guardRequest) (*http.Response, error) {
	return outboundHandler.Get(req.ctx, req.name, req.url)
}

// HttpPost is a fiberstanza helper function (passthrough to stanza.NewHttpOutboundHandler)
func HttpPost(req guardRequest) (*http.Response, error) {
	return outboundHandler.Post(req.ctx, req.name, req.url, req.body)
}

// Guard is a fiberstanza helper function
func Guard(c *fiber.Ctx, name string, url string, opts ...Opt) guardRequest {
	var req http.Request
	fasthttpadaptor.ConvertRequest(c.Context(), &req, true)
	ctx := otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
	return guardRequest{ctx: addKeys(ctx, opts...), name: name, url: url}
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
