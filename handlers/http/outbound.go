package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/StanzaSystems/sdk-go/keys"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type OutboundHandler struct {
	apikey          string
	clientId        string
	customerId      string
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

func NewOutboundHandler(apikey, clientId, environment, service string, otelEnabled, sentinelEnabled bool) (*OutboundHandler, error) {
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
			clientIdKey.String(clientId),
			environmentKey.String(environment),
			serviceKey.String(service),
		},
	}
	if m, err := GetMeter(); err != nil {
		return nil, err
	} else {
		handler.meter = m
		return handler, nil
	}
}

func (h *OutboundHandler) Meter() *Meter {
	return h.meter
}

func (h *OutboundHandler) SetCustomerId(id string) {
	if h.customerId == "" {
		h.customerId = id
		h.attr = append(h.attr, customerIdKey.String(id))
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
	return h.Request(ctx, http.MethodGet, url, http.NoBody, tlr)
}

func (h *OutboundHandler) Post(ctx context.Context, url string, body io.Reader, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	return h.Request(ctx, http.MethodPost, url, body, tlr)
}

func (h *OutboundHandler) Request(ctx context.Context, httpMethod, url string, body io.Reader, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	// Inspect Baggage and Headers for Feature and PriorityBoost, propagate through context if found
	ctx, tlr.Selector.FeatureName = getFeature(ctx, tlr.Selector.GetFeatureName())
	ctx, tlr.PriorityBoost = getPriorityBoost(ctx, tlr.GetPriorityBoost())

	// Add Decorator and Feature to OTEL attributes
	attr := append(h.attr,
		decoratorKey.String(tlr.Selector.GetDecoratorName()),
		featureKey.String(tlr.Selector.GetFeatureName()),
	)

	// TODO: Add a Span around this Request, like otelhttp does:
	// https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp#Get
	tlr.ClientId = &h.clientId
	tlr.Selector.Environment = h.environment
	if req, err := http.NewRequestWithContext(ctx, httpMethod, url, body); err != nil {
		h.meter.FailedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
		return nil, err
	} else {
		start := time.Now()
		resp, err := h.request(ctx, req, tlr, attr)
		if err != nil {
			h.meter.FailedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
		} else {
			h.meter.SucceededCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
		}
		recAttr := []metric.RecordOption{metric.WithAttributes(append(attr,
			httpRequestMethodKey.String(httpMethod),
			httpResponseCodeKey.Int(resp.StatusCode))...)}
		h.meter.ClientRequestSize.Record(ctx, resp.ContentLength, recAttr...)
		h.meter.ClientResponseSize.Record(ctx, req.ContentLength, recAttr...)
		h.meter.ClientDuration.Record(ctx, float64(time.Since(start).Microseconds())/1000, recAttr...)
		return resp, err
	}
}

func (h *OutboundHandler) request(ctx context.Context, req *http.Request, tlr *hubv1.GetTokenLeaseRequest, attr []attribute.KeyValue) (*http.Response, error) {
	if ok, token := checkQuota(h.apikey, h.decoratorConfig[tlr.Selector.DecoratorName], h.qsc, tlr); ok {
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
		h.meter.AllowedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
		httpClient := &http.Client{
			Transport: otelhttp.NewTransport(
				http.DefaultTransport,
			)}
		return httpClient.Do(req)
	} else {
		h.meter.BlockedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
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
