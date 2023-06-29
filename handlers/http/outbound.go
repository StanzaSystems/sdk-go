package http

import (
	"context"
	"fmt"
	"io"
	"net/http"

	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// const (
// 	// Standard HTTP Client Metrics:
// 	// https://opentelemetry.io/docs/specs/otel/metrics/semantic_conventions/http-metrics/#http-client
// 	httpClientDuration     = "http.client.duration"      // histogram
// 	httpClientRequestSize  = "http.client.request.size"  // histogram
// 	httpClientResponseSize = "http.client.response.size" // histogram
// )

type OutboundHandler struct {
	apikey          string
	clientId        string
	environment     string
	otelEnabled     bool
	sentinelEnabled bool

	decoratorConfig map[string]*hubv1.DecoratorConfig
	qsc             hubv1.QuotaServiceClient
	propagators     propagation.TextMapPropagator
	tracer          trace.Tracer
	meter           *Meter
	attr            []attribute.KeyValue
}

func NewOutboundHandler(apikey, environment, clientId string, otelEnabled, sentinelEnabled bool) (*OutboundHandler, error) {
	handler := &OutboundHandler{
		apikey:          apikey,
		clientId:        clientId,
		environment:     environment,
		otelEnabled:     otelEnabled,
		sentinelEnabled: sentinelEnabled,
		decoratorConfig: make(map[string]*hubv1.DecoratorConfig),
		qsc:             nil,
		propagators:     otel.GetTextMapPropagator(),
		tracer: otel.GetTracerProvider().Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(instrumentationVersion),
		),
		attr: []attribute.KeyValue{
			// CustomerID
			clientIdKey.String(clientId),
			environmentKey.String(environment),
		},
	}
	if m, err := GetMeter(); err != nil {
		return nil, err
	} else {
		handler.meter = m
		return handler, nil
	}
}

func (h *OutboundHandler) SetDecoratorConfig(d string, dc *hubv1.DecoratorConfig) {
	if h.decoratorConfig[d] == nil {
		h.decoratorConfig[d] = dc
	}
}

func (h *OutboundHandler) SetQuotaServiceClient(quotaServiceClient hubv1.QuotaServiceClient) {
	if h.qsc == nil {
		h.qsc = quotaServiceClient
	}
}

func (h *OutboundHandler) Get(ctx context.Context, url string, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	tlr.ClientId = &h.clientId
	tlr.Selector.Environment = h.environment
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody); err != nil {
		return nil, err
	} else {
		return h.Request(ctx, req, tlr)
	}
}

func (h *OutboundHandler) Post(ctx context.Context, url string, body io.Reader, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	tlr.ClientId = &h.clientId
	tlr.Selector.Environment = h.environment
	if req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body); err != nil {
		return nil, err
	} else {
		return h.Request(ctx, req, tlr)
	}
}

func (h *OutboundHandler) Request(ctx context.Context, req *http.Request, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	attr := append(h.attr,
		decoratorKey.String(tlr.Selector.GetDecoratorName()),
		featureKey.String(tlr.Selector.GetFeatureName()),
	)
	h.meter.TotalCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
	headers := ctx.Value("StanzaOutboundHeaders")
	if headers != nil {
		for k, v := range headers.(map[string]string) {
			req.Header.Set(k, v)
		}
	}
	if ok, token := checkQuota(h.apikey, h.decoratorConfig[tlr.Selector.DecoratorName], h.qsc, tlr); ok {
		if token != "" {
			req.Header.Add("X-Stanza-Token", token)
		}
		return http.DefaultClient.Do(req)
	} else {
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
