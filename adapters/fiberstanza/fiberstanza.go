package fiberstanza

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/StanzaSystems/sdk-go/keys"
	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/StanzaSystems/sdk-go/stanza"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

// Client defines options for a new Stanza Client
type Client struct {
	// Required
	APIKey string // customer generated API key

	// Optional
	Name        string   // defines applications name
	Release     string   // defines applications version
	Environment string   // defines applications environment
	StanzaHub   string   // host:port (ipv4, ipv6, or resolvable hostname)
	Guard       []string // prefetch config for these guards
}

// Optional arguments
type Opt struct {
	Headers       http.Header
	Feature       string
	PriorityBoost int32
	DefaultWeight float32
}

// New creates a new fiberstanza middleware fiber.Handler
func New(guardName string, opts ...Opt) fiber.Handler {
	h, err := stanza.HttpServer(guardName, withOpts(opts...))
	if err != nil {
		logging.Error(fmt.Errorf("failed to create HTTP inbound handler: %v", err))
		return func(c *fiber.Ctx) error {
			// with no InboundHandler there is nothing we can do but fail open
			logging.Error(fmt.Errorf("no HTTP inbound handler, failing open"))
			if h != nil {
				h.FailOpen(c.UserContext())
			}
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		if h == nil {
			// with no InboundHandler there is nothing we can do but fail open
			logging.Error(fmt.Errorf("no HTTP inbound handler, failing open"))
			return c.Next()
		}

		var req http.Request
		if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
			// if we can't convert from fasthttp to http.Request, log the error and fail open
			logging.Error(fmt.Errorf("failed to convert request from fasthttp: %v", err))
			if h != nil {
				h.FailOpen(c.UserContext())
			}
			return c.Next()
		}

		ctx, span, tokens := h.Start(&req)
		defer span.End()

		guard := h.Guard(ctx, span, tokens)
		c.SetUserContext(guard.Context())

		// Stanza Blocked
		if guard.Blocked() {
			span.SetAttributes(semconv.HTTPStatusCode(http.StatusTooManyRequests))
			span.SetStatus(codes.Error, guard.BlockMessage())
			c.SendString(guard.BlockMessage())
			return c.SendStatus(http.StatusTooManyRequests)
		}

		// Stanza Allowed
		err := c.Next() // intercept c.Next() for guard.End() status
		span.SetAttributes(semconv.HTTPStatusCode(c.Response().StatusCode()))
		span.SetStatus(h.HTTPServerStatus(c.Response().StatusCode()))
		if err != nil {
			span.RecordError(err)
			// invokes the registered HTTP error handler
			// to get the correct response status code
			_ = c.App().Config().ErrorHandler(c, err)
			guard.End(guard.Failure)
		} else {
			guard.End(guard.Success)
		}
		return nil
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
func HttpGet(c *fiber.Ctx, guardName string, url string, opts ...Opt) (*http.Response, error) {
	var req http.Request
	fasthttpadaptor.ConvertRequest(c.Context(), &req, true)
	ctx := otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
	return stanza.HttpGet(withHeaders(ctx, opts...), guardName, url, withOpts(opts...))
}

// HttpPost is a fiberstanza helper function (passthrough to stanza.HttpPost)
func HttpPost(c *fiber.Ctx, guardName string, url string, body io.Reader, opts ...Opt) (*http.Response, error) {
	var req http.Request
	fasthttpadaptor.ConvertRequest(c.Context(), &req, true)
	ctx := otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
	return stanza.HttpPost(withHeaders(ctx, opts...), guardName, url, body, withOpts(opts...))
}

func withHeaders(ctx context.Context, opts ...Opt) context.Context {
	if len(opts) == 1 {
		if opts[0].Headers != nil {
			ctx = context.WithValue(ctx, keys.OutboundHeadersKey, opts[0].Headers)
		}
	}
	return ctx
}

func withOpts(opts ...Opt) stanza.GuardOpt {
	guardOpt := stanza.GuardOpt{}
	if len(opts) == 1 {
		if opts[0].Feature != "" {
			guardOpt.Feature = &opts[0].Feature
		}
		if opts[0].PriorityBoost != 0 {
			guardOpt.PriorityBoost = &opts[0].PriorityBoost
		}
		if opts[0].DefaultWeight != 0 {
			guardOpt.DefaultWeight = &opts[0].DefaultWeight
		}
	}
	return guardOpt
}
