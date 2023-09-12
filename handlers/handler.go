package handlers

import (
	"context"
	"time"

	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/hub"
	"github.com/StanzaSystems/sdk-go/otel"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	guardName     string
	featureName   *string // overrides request baggage (if any)
	priorityBoost *int32  // adds to request baggage (if any)
	defaultWeight *float32
	tracer        trace.Tracer
	meter         *Meter
	attr          []attribute.KeyValue
}

func NewHandler(gn string, fn *string, pb *int32, dw *float32) (*Handler, error) {
	m, err := GetStanzaMeter()
	return &Handler{
		guardName:     gn,
		featureName:   fn,
		priorityBoost: pb,
		defaultWeight: dw,
		meter:         m,
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

func (h *Handler) Guard(ctx context.Context, span trace.Span, tokens []string) *Guard {
	if span == nil {
		// Default OTEL Tracer if none specified
		opts := []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindUnspecified),
		}
		ctx, span = h.Tracer().Start(ctx, h.GuardName(), opts...)
		defer span.End()
	}

	ctx, tlr := hub.NewTokenLeaseRequest(ctx, h.GuardName(), h.FeatureName(), h.PriorityBoost(), h.DefaultWeight())
	attr := []attribute.KeyValue{
		guardKey.String(tlr.Selector.GetGuardName()),
		featureKey.String(tlr.Selector.GetFeatureName()),
	}

	g := h.NewGuard(ctx, span, attr, nil)
	if h.SentinelEnabled() {
		g.checkSentinel(h.GuardName())
		if g.sentinelBlock != nil {
			return g
		}
	}
	if len(tokens) > 0 {
		g.checkToken(ctx, h.GuardName(), tokens)
		if g.quotaStatus == hub.ValidateTokensInvalid {
			return g
		}
	}
	g.checkQuota(ctx, tlr)
	g.start = time.Now()
	return g
}

func (h *Handler) NewGuard(ctx context.Context, span trace.Span, attr []attribute.KeyValue, err error) *Guard {
	attr = append(attr, customerIdKey.String(global.GetCustomerID()))
	return &Guard{
		ctx:   ctx,
		start: time.Time{},
		meter: h.meter,
		span:  span,
		attr:  append(h.attr, attr...),
		err:   err,

		Success:     GuardSuccess,
		Failure:     GuardFailure,
		Unknown:     GuardUnknown,
		finalStatus: GuardUnknown,

		sentinelBlock: nil,
	}
}

func (h *Handler) GuardName() string {
	return h.guardName
}

func (h *Handler) FeatureName() *string {
	return h.featureName
}

func (h *Handler) PriorityBoost() *int32 {
	return h.priorityBoost
}

func (h *Handler) DefaultWeight() *float32 {
	return h.defaultWeight
}

// OTEL Helper Functions //
func (h *Handler) Tracer() trace.Tracer {
	return h.tracer
}

func (h *Handler) Propagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

func (h *Handler) FailOpen(ctx context.Context) {
	if h.meter != nil {
		h.meter.AllowedCount.Add(ctx, 1,
			[]metric.AddOption{metric.WithAttributes(append(h.attr,
				reason(ReasonFailOpen))...)}...)
		h.meter.AllowedUnknownCount.Add(ctx, 1,
			[]metric.AddOption{metric.WithAttributes(h.attr...)}...)
	}
}

// Global Helper Functions //
func (h *Handler) OTELEnabled() bool {
	return global.OtelEnabled()
}

func (h *Handler) SentinelEnabled() bool {
	return global.SentinelEnabled()
}
