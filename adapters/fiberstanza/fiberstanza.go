package fiberstanza

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/StanzaSystems/sdk-go/handlers/httphandler"
	"github.com/StanzaSystems/sdk-go/keys"
	"github.com/StanzaSystems/sdk-go/logging"
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
func New(guard string, opts ...Opt) fiber.Handler {
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
		// h.SetTokenLeaseRequest(guard, tokenLeaseRequest(guard, opts...))
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
		ctx, status := h.VerifyServingCapacity(&req, c.Route().Path, guard)
		if status != http.StatusOK {
			c.SendString("Stanza Inbound Rate Limited")
			return c.SendStatus(status)
		}
		c.SetUserContext(ctx)

		start := time.Now()
		savedCtx, cancel := context.WithCancel(c.UserContext())
		defer func() {
			h.Meter().AllowedDuration.Record(savedCtx,
				float64(time.Since(start).Microseconds())/1000,
				[]metric.RecordOption{metric.WithAttributes(h.Attributes()...)}...)
			c.SetUserContext(savedCtx)
			cancel()
		}()
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
	return guardRequest{ctx: ctx, name: name, url: url}
}
