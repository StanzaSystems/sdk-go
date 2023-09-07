package handlers

import (
	"context"
	"time"

	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/hub"
	"github.com/StanzaSystems/sdk-go/otel"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	tracer trace.Tracer
	meter  *Meter
	attr   []attribute.KeyValue
}

func NewHandler() (*Handler, error) {
	m, err := GetStanzaMeter()
	return &Handler{
		meter: m,
		tracer: otel.GetTracerProvider().Tracer(
			global.InstrumentationName(),
			global.InstrumentationTraceVersion(),
		),
		attr: []attribute.KeyValue{
			clientIdKey.String(global.GetClientID()),
			environmentKey.String(global.GetServiceEnvironment()),
			serviceKey.String(global.GetServiceName()),
		},
	}, err
}

func (h *Handler) NewGuard(ctx context.Context, span trace.Span, name string, tokens []string) *Guard {
	if span == nil {
		// Default OTEL Tracer if none specified
		opts := []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindUnspecified),
		}
		ctx, span = h.Tracer().Start(ctx, name, opts...)
		defer span.End()
	}

	tlr := hub.NewTokenLeaseRequest(ctx, name)
	attr := []attribute.KeyValue{
		h.GuardKey(tlr.Selector.GetGuardName()),
		h.FeatureKey(tlr.Selector.GetFeatureName()),
	}

	g := h.NewGuardError(ctx, span, attr, nil)
	if h.SentinelEnabled() {
		g.checkSentinel(name)
		if g.sentinelBlock != nil {
			return g
		}
	}
	if len(tokens) > 0 {
		g.checkToken(ctx, name, tokens)
		if g.quotaStatus == hub.ValidateTokensInvalid {
			return g
		}
	}
	g.checkQuota(ctx, tlr)
	g.start = time.Now()
	return g
}

func (h *Handler) NewGuardError(ctx context.Context, span trace.Span, attr []attribute.KeyValue, err error) *Guard {
	return &Guard{
		ctx:   ctx,
		start: time.Time{},
		meter: h.meter,
		span:  span,
		attr:  append(h.Attributes(), attr...),
		err:   err,

		Success:     GuardSuccess,
		Failure:     GuardFailure,
		Unknown:     GuardUnknown,
		finalStatus: GuardUnknown,

		sentinelBlock: nil,
	}
}

func (h *Handler) Meter() *Meter {
	return h.meter
}

func (h *Handler) Tracer() trace.Tracer {
	return h.tracer
}

func (h *Handler) Propagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

// OTEL Attribute Helper Functions //
func (h *Handler) Attributes() []attribute.KeyValue {
	return append(h.attr, customerIdKey.String(global.GetCustomerID()))
}

func (h *Handler) FeatureKey(feat string) attribute.KeyValue {
	return featureKey.String(feat)
}

func (h *Handler) GuardKey(guard string) attribute.KeyValue {
	return guardKey.String(guard)
}

func (h *Handler) HttpStatusCodeKey(code int) attribute.KeyValue {
	return httpStatusCodeKey.Int(code)
}

func (h *Handler) HttpUserAgentKey(ua string) attribute.KeyValue {
	return httpUserAgentKey.String(ua)
}

func (h *Handler) ReasonFailOpen() attribute.KeyValue {
	return reason(ReasonFailOpen)
}

// Global Helper Functions //
func (h *Handler) APIKey() string {
	return global.GetServiceKey()
}

func (h *Handler) ClientID() string {
	return global.GetClientID()
}

func (h *Handler) Environment() string {
	return global.GetServiceEnvironment()
}

func (h *Handler) OTELEnabled() bool {
	return global.OtelEnabled()
}

func (h *Handler) SentinelEnabled() bool {
	return global.SentinelEnabled()
}

func (h *Handler) QuotaServiceClient() hubv1.QuotaServiceClient {
	return global.QuotaServiceClient()
}
