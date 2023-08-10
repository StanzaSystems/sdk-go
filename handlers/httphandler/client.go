package httphandler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/hub"
	"github.com/StanzaSystems/sdk-go/keys"
	"github.com/StanzaSystems/sdk-go/otel"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/protobuf/proto"
)

type OutboundHandler struct {
	*handlers.OutboundHandler
}

// NewOutboundHandler returns a new OutboundHandler
func NewOutboundHandler(apikey, clientId, environment, service string, otelEnabled, sentinelEnabled bool) (*OutboundHandler, error) {
	h, err := handlers.NewOutboundHandler(apikey, clientId, environment, service, otelEnabled, sentinelEnabled)
	if err != nil {
		return nil, err
	}
	return &OutboundHandler{h}, nil
}

// Get wraps a HTTP GET request
func (h *OutboundHandler) Get(ctx context.Context, url string, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	return h.Request(ctx, http.MethodGet, url, http.NoBody, tlr)
}

// Post wraps a HTTP POST request
func (h *OutboundHandler) Post(ctx context.Context, url string, body io.Reader, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	return h.Request(ctx, http.MethodPost, url, body, tlr)
}

// Request wraps a HTTP request of the given HTTP method
func (h *OutboundHandler) Request(ctx context.Context, httpMethod, url string, body io.Reader, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	// Inspect Baggage and Headers for Feature and PriorityBoost, propagate through context if found
	ctx, tlr.Selector.FeatureName = otel.GetFeature(ctx, tlr.Selector.GetFeatureName())
	ctx, tlr.PriorityBoost = otel.GetPriorityBoost(ctx, tlr.GetPriorityBoost())

	// Add Decorator and Feature to OTEL attributes
	attr := append(h.Attributes(),
		h.DecoratorKey(tlr.Selector.GetDecoratorName()),
		h.FeatureKey(tlr.Selector.GetFeatureName()),
	)

	// TODO: Add a Span around this Request, like otelhttp does:
	// https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp#Get
	tlr.ClientId = proto.String(h.ClientID())
	tlr.Selector.Environment = h.Environment()
	if req, err := http.NewRequestWithContext(ctx, httpMethod, url, body); err != nil {
		h.StanzaMeter().AllowedFailureCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
		return nil, err
	} else {
		start := time.Now()
		resp, err := h.request(ctx, req, tlr, attr)
		if err != nil {
			h.StanzaMeter().AllowedFailureCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
		} else {
			h.StanzaMeter().AllowedSuccessCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
		}
		recAttr := []metric.RecordOption{metric.WithAttributes(attr...)}
		h.StanzaMeter().AllowedDuration.Record(ctx, float64(time.Since(start).Microseconds())/1000, recAttr...)
		return resp, err
	}
}

func (h *OutboundHandler) request(ctx context.Context, req *http.Request, tlr *hubv1.GetTokenLeaseRequest, attr []attribute.KeyValue) (*http.Response, error) {
	if ok, token := hub.CheckQuota(
		h.APIKey(),
		h.DecoratorConfig(tlr.Selector.DecoratorName),
		h.QuotaServiceClient(),
		tlr); ok {
		if token != "" {
			req.Header.Add("X-Stanza-Token", token)
		}

		if req.Header.Get("User-Agent") == "" {
			req.Header.Set("User-Agent", "StanzaGoSDK/v0.0.1-beta") // TODO: Prefix with Service/Release
		}
		if ctx.Value(keys.OutboundHeadersKey) != nil {
			for k, v := range ctx.Value(keys.OutboundHeadersKey).(http.Header) {
				req.Header.Set(k, v[0])
			}
		}
		h.StanzaMeter().AllowedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
		httpClient := &http.Client{
			Transport: otelhttp.NewTransport(
				http.DefaultTransport,
			)}
		return httpClient.Do(req)
	} else {
		h.StanzaMeter().BlockedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
		return &http.Response{
			Status:     fmt.Sprintf("%d Too Many Request", http.StatusTooManyRequests),
			StatusCode: http.StatusTooManyRequests,
			Request:    req,
			Body:       http.NoBody,
			Header:     http.Header{
				// TODO: Add retry-after header
			},
		}, nil
	}
}
