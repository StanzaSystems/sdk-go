package handlers

import (
	"context"
	"time"

	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/hub"
	"github.com/StanzaSystems/sdk-go/keys"
	"github.com/StanzaSystems/sdk-go/otel"
	"google.golang.org/protobuf/proto"

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

	g := h.NewGuardError(span, nil)
	tlr := hub.NewTokenLeaseRequest(name)

	// Inspect Baggage and Headers for Feature and PriorityBoost,
	// propagate through context if found
	ctx, tlr.Selector.FeatureName = otel.GetFeature(ctx, tlr.Selector.GetFeatureName())
	ctx, tlr.PriorityBoost = otel.GetPriorityBoost(ctx, tlr.GetPriorityBoost())

	// Override Baggage and Header values with user supplied data.
	if ctx.Value(keys.StanzaFeatureNameKey) != nil {
		tlr.Selector.FeatureName = proto.String(ctx.Value(keys.StanzaFeatureNameKey).(string))
		ctx, tlr.Selector.FeatureName = otel.GetFeature(ctx, tlr.Selector.GetFeatureName())
	}
	if ctx.Value(keys.StanzaPriorityBoostKey) != nil {
		tlr.PriorityBoost = proto.Int32(ctx.Value(keys.StanzaPriorityBoostKey).(int32))
		ctx, tlr.PriorityBoost = otel.GetPriorityBoost(ctx, tlr.GetPriorityBoost())
	}
	if ctx.Value(keys.StanzaDefaultWeightKey) != nil {
		tlr.DefaultWeight = proto.Float32(ctx.Value(keys.StanzaDefaultWeightKey).(float32))
	}
	g.ctx = ctx

	// Add Guard and Feature to OTEL attributes
	g.attr = append(h.Attributes(),
		h.GuardKey(tlr.Selector.GetGuardName()),
		h.FeatureKey(tlr.Selector.GetFeatureName()),
	)

	// Check Sentinel
	if h.SentinelEnabled() {
		g.checkSentinel(name)
	}

	// Check Quota Token ()
	if len(tokens) > 0 {
		g.checkToken(ctx, name, tokens)
	}

	// Check Quota
	g.checkQuota(ctx, tlr)

	g.start = time.Now()
	return g
}

func (h *Handler) NewGuardError(span trace.Span, err error) *Guard {
	return &Guard{
		Success:       GuardSuccess,
		Failure:       GuardFailure,
		Unknown:       GuardUnknown,
		finalStatus:   GuardUnknown,
		err:           err,
		span:          span,
		meter:         h.meter,
		quotaMessage:  "",
		quotaToken:    "",
		quotaReason:   "quota_unknown",
		quotaStatus:   GuardUnknown,
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

func (h *Handler) GuardKey(guard string) attribute.KeyValue {
	return guardKey.String(guard)
}

func (h *Handler) FeatureKey(feat string) attribute.KeyValue {
	return featureKey.String(feat)
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
