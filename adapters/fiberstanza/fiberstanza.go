package fiberstanza

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/StanzaSystems/sdk-go/handlers/httphandler"
	"github.com/StanzaSystems/sdk-go/keys"
	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/stanza"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
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
}

// Optional arguments
type Opt struct {
	Headers       http.Header
	Feature       string
	PriorityBoost int32
	DefaultWeight float32
}

type decorateRequest struct {
	c       *fiber.Ctx
	tlr     *hubv1.GetTokenLeaseRequest
	headers http.Header
	url     string
	body    io.Reader
}

var (
	inboundHandler  *httphandler.InboundHandler  = nil
	outboundHandler *httphandler.OutboundHandler = nil
	seenDecorators  map[string]bool              = make(map[string]bool)
)

// New creates a new fiberstanza middleware fiber.Handler
func New(decorator string, opts ...Opt) fiber.Handler {
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
		h.SetTokenLeaseRequest(decorator, tokenLeaseRequest(decorator, opts...))
		inboundHandler = h
	}
	h := inboundHandler

	return func(c *fiber.Ctx) error {
		var req http.Request
		addAttr := []metric.AddOption{metric.WithAttributes(h.Attributes()...)}
		if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
			logging.Error(fmt.Errorf("failed to convert request from fasthttp: %v", err))
			// TODO:ADD "reason = fail_open"
			h.StanzaMeter().AllowedSuccessCount.Add(c.UserContext(), 1, addAttr...)
			return c.Next() // log error and fail open
		}
		ctx, status := h.VerifyServingCapacity(&req, c.Route().Path, decorator)
		if status != http.StatusOK {
			c.SendString("Stanza Inbound Rate Limited")
			return c.SendStatus(status)
		}
		c.SetUserContext(ctx)
		return c.Next()
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
func HttpGet(req decorateRequest) (*http.Response, error) {
	return outboundHandler.Get(WithHeaders(req.c, req.headers), req.url, req.tlr)
}

// HttpPost is a fiberstanza helper function (passthrough to stanza.NewHttpOutboundHandler)
func HttpPost(req decorateRequest) (*http.Response, error) {
	return outboundHandler.Post(WithHeaders(req.c, req.headers), req.url, req.body, req.tlr)
}

// Add Headers to Context
func WithHeaders(c *fiber.Ctx, headers http.Header) context.Context {
	var req http.Request
	if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
		logging.Error(fmt.Errorf("failed to convert request from fasthttp: %v", err))
	}
	ctx := otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
	return context.WithValue(ctx, keys.OutboundHeadersKey, headers)
}

// Decorate is a fiberstanza helper function
func Decorate(c *fiber.Ctx, decorator string, url string, opts ...Opt) decorateRequest {
	req := decorateRequest{c: c, headers: make(http.Header)}
	tlr := tokenLeaseRequest(decorator, opts...)
	if len(opts) == 1 {
		if opts[0].Headers != nil {
			req.headers = opts[0].Headers
		}
	}
	req.tlr = tlr
	req.url = url
	return req
}

func tokenLeaseRequest(decorator string, opts ...Opt) *hubv1.GetTokenLeaseRequest {
	if _, ok := seenDecorators[decorator]; !ok {
		stanza.GetDecoratorConfig(context.Background(), decorator)
		seenDecorators[decorator] = true
	}
	dfs := &hubv1.DecoratorFeatureSelector{DecoratorName: decorator}
	tlr := &hubv1.GetTokenLeaseRequest{Selector: dfs}
	if len(opts) == 1 {
		if opts[0].Feature != "" {
			tlr.Selector.FeatureName = &opts[0].Feature
		}
		if opts[0].PriorityBoost != 0 {
			tlr.PriorityBoost = &opts[0].PriorityBoost
		}
		if opts[0].DefaultWeight != 0 {
			tlr.DefaultWeight = &opts[0].DefaultWeight
		}
	}
	return tlr
}
