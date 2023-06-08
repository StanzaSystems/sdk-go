package fiberstanza

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/stanza"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.opentelemetry.io/otel/metric"
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

// New creates a new fiberstanza middleware fiber.Handler
func Middleware(ctx context.Context, decorator string) fiber.Handler {
	h, err := stanza.NewHttpInboundHandler(ctx, decorator)
	if err != nil {
		logging.Error(fmt.Errorf("failed to initialize new http inbound meters: %v", err))
	}

	return func(c *fiber.Ctx) error {
		start := time.Now()
		savedCtx, cancel := context.WithCancel(c.UserContext())

		addAttr := []metric.AddOption{metric.WithAttributes(h.Meter().Attributes...)}
		recAttr := []metric.RecordOption{metric.WithAttributes(h.Meter().Attributes...)}
		h.Meter().ActiveRequests.Add(savedCtx, 1, addAttr...)
		defer func() {
			h.Meter().Duration.Record(savedCtx, float64(time.Since(start).Microseconds())/1000, recAttr...)
			h.Meter().RequestSize.Record(savedCtx, int64(len(c.Request().Body())), recAttr...)
			h.Meter().ResponseSize.Record(savedCtx, int64(len(c.Response().Body())), recAttr...)
			h.Meter().ActiveRequests.Add(savedCtx, -1, addAttr...)
			c.SetUserContext(savedCtx)
			cancel()
		}()

		// TODO(msg): implement HttpInboundHandler as fasthttp handler instead of converting to net/http?
		var req http.Request
		if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
			logging.Error(fmt.Errorf("failed to convert request from fasthttp: %v", err))
			return c.Next() // log error and fail open
		}
		ctx, status := h.VerifyServingCapacity(&req, c.Route().Path)
		if status != http.StatusOK {
			return c.SendStatus(status)
		}
		c.SetUserContext(ctx)
		return c.Next()
	}
}

// Init is a fiberstanza helper function (passthrough to stanza.Init)
func Init(ctx context.Context, client Client) (func(), error) {
	return stanza.Init(ctx, stanza.ClientOptions(client))
}

// HttpGet is a fiberstanza helper function (passthrough to stanza.NewHttpOutboundHandler)
func HttpGet(ctx context.Context, url string, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	return stanza.NewHttpOutboundHandler(ctx, http.MethodGet, url, http.NoBody, tlr)
}

// HttpGet is a fiberstanza helper function (passthrough to stanza.NewHttpOutboundHandler)
func HttpPost(ctx context.Context, url string, body io.Reader, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	return stanza.NewHttpOutboundHandler(ctx, http.MethodPost, url, body, tlr)
}

type Opt struct {
	PriorityBoost int32
	DefaultWeight float32
}

// Decorate is a fiberstanza helper function
func Decorate(decorator string, feature string, opts ...Opt) *hubv1.GetTokenLeaseRequest {
	dfs := &hubv1.DecoratorFeatureSelector{DecoratorName: decorator}
	if feature != "" {
		dfs.FeatureName = &feature
	}
	tlr := &hubv1.GetTokenLeaseRequest{Selector: dfs}
	if len(opts) == 1 {
		if opts[0].PriorityBoost != 0 {
			tlr.PriorityBoost = &opts[0].PriorityBoost
		}
		if opts[0].DefaultWeight != 0 {
			tlr.DefaultWeight = &opts[0].DefaultWeight
		}
	}
	return tlr
}

// GetFeatureFromContext is a helper function to extract stanza feature name from
// OTEL baggage (which is hiding in the fiber.Ctx)
func GetFeatureFromContext(c *fiber.Ctx) string {
	// TODO: actually extract STANZA_FEATURE from OTEL Baggage
	//
	// var req http.Request	//
	//	if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
	//		logging.Error(fmt.Errorf("failed to convert request from fasthttp: %v", err))
	//	}
	// ctx := otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
	// return "FOOBAR"
	return ""
}

func NoFeature() string {
	return ""
}
