package http

import (
	"context"
	"net/http"

	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

type InboundHandler struct {
	apikey          string
	clientId        string
	customerId      string
	environment     string
	otelEnabled     bool
	sentinelEnabled bool

	decoratorConfig map[string]*hubv1.DecoratorConfig
	tlr             map[string]*hubv1.GetTokenLeaseRequest
	qsc             hubv1.QuotaServiceClient
	propagators     propagation.TextMapPropagator
	tracer          trace.Tracer
	meter           *Meter
	attr            []attribute.KeyValue
}

// New returns a new InboundHandler
func NewInboundHandler(apikey, clientId, environment, service string, otelEnabled, sentinelEnabled bool) (*InboundHandler, error) {
	handler := &InboundHandler{
		apikey:          apikey,
		clientId:        clientId,
		environment:     environment,
		otelEnabled:     otelEnabled,
		sentinelEnabled: sentinelEnabled,
		decoratorConfig: make(map[string]*hubv1.DecoratorConfig),
		tlr:             make(map[string]*hubv1.GetTokenLeaseRequest),
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

func (h *InboundHandler) Attributes() []attribute.KeyValue {
	return h.attr
}

func (h *InboundHandler) Meter() *Meter {
	return h.meter
}

func (h *InboundHandler) SetCustomerId(id string) {
	if h.customerId == "" {
		h.customerId = id
		h.attr = append(h.attr, customerIdKey.String(id))
	}
}

func (h *InboundHandler) SetDecoratorConfig(d string, dc *hubv1.DecoratorConfig) {
	if h.decoratorConfig[d] == nil {
		h.decoratorConfig[d] = dc
	}
}

func (h *InboundHandler) SetQuotaServiceClient(quotaServiceClient hubv1.QuotaServiceClient) {
	if h.qsc == nil {
		h.qsc = quotaServiceClient
	}
}

func (h *InboundHandler) SetTokenLeaseRequest(d string, tlr *hubv1.GetTokenLeaseRequest) {
	tlr.ClientId = &h.clientId
	tlr.Selector.Environment = h.environment
	if h.tlr[d] == nil {
		h.tlr[d] = tlr
	}
}

func (h *InboundHandler) VerifyServingCapacity(r *http.Request, route string, decorator string) (context.Context, int) {
	ctx := h.propagators.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	tlr := proto.Clone(h.tlr[decorator]).(*hubv1.GetTokenLeaseRequest)

	// Inspect Baggage and Headers for Feature and PriorityBoost, propagate through context if found
	ctx, tlr.Selector.FeatureName = getFeature(ctx, tlr.Selector.GetFeatureName())
	ctx, tlr.PriorityBoost = getPriorityBoost(ctx, tlr.GetPriorityBoost())

	// Add Decorator and Feature to OTEL attributes
	attr := append(h.attr,
		decoratorKey.String(tlr.Selector.GetDecoratorName()),
		featureKey.String(tlr.Selector.GetFeatureName()),
	)

	// generic HTTP server trace
	if route == "" {
		route = r.URL.Path
	}
	opts := []trace.SpanStartOption{
		trace.WithAttributes(httpconv.ServerRequest("", r)...),
		trace.WithSpanKind(trace.SpanKindServer),
	}
	ctx, span := h.tracer.Start(ctx, route, opts...)
	defer span.End()

	if h.sentinelEnabled {
		e, b := api.Entry(decorator, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
		if b != nil {
			logging.Debug("Stanza blocked",
				"decorator", decorator,
				"sentinel.block_msg", b.BlockMsg(),
				"sentinel.block_type", b.BlockType().String(),
				"sentinel.block_value", b.TriggeredValue(),
				"sentinel.block_rule", b.TriggeredRule().String(),
			)
			attrWithReason := append(attr, reasonKey.String(b.BlockType().String()))
			span.AddEvent("Stanza blocked", trace.WithAttributes(attrWithReason...))
			h.meter.BlockedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
			return ctx, http.StatusTooManyRequests
		}
		e.Exit() // cleanly exit the Sentinel Entry
	}

	if ok, _ := checkQuota(h.apikey, h.decoratorConfig[decorator], h.qsc, tlr); !ok {
		attrWithReason := append(attr, reasonKey.String("quota"))
		span.AddEvent("Stanza blocked", trace.WithAttributes(attrWithReason...))
		h.meter.BlockedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
		return ctx, http.StatusTooManyRequests
	}

	span.AddEvent("Stanza allowed", trace.WithAttributes(attr...))
	h.meter.AllowedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
	return ctx, http.StatusOK // return success
}
