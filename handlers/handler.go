package handlers

import (
	"context"

	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/hub"
	"github.com/StanzaSystems/sdk-go/keys"
	"github.com/StanzaSystems/sdk-go/otel"
	"google.golang.org/protobuf/proto"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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

func (h *Handler) NewGuard(ctx context.Context, name string) *Guard {
	g := h.NewGuardError(nil)
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

	// Check Quota
	g.quotaStatus, g.quotaToken = hub.CheckQuota(ctx, tlr)

	// Add Guard and Feature to OTEL attributes
	attr := append(h.Attributes(),
		h.GuardKey(tlr.Selector.GetGuardName()),
		h.FeatureKey(tlr.Selector.GetFeatureName()),
	)
	switch g.quotaStatus {
	case hub.CheckQuotaBlocked:
		attr = append(attr, h.ReasonQuota())
		h.Meter().BlockedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
		g.quotaReason = h.ReasonQuota().Value.AsString()
		return g
	case hub.CheckQuotaAllowed:
		attr = append(attr, h.ReasonQuota())
		g.quotaReason = h.ReasonQuota().Value.AsString()
	case hub.CheckQuotaSkipped:
		attr = append(attr, h.ReasonQuotaCheckDisabled())
		g.quotaReason = h.ReasonQuotaCheckDisabled().Value.AsString()
	case hub.CheckQuotaFailOpen:
		attr = append(attr, h.ReasonQuotaFailOpen())
		g.quotaReason = h.ReasonQuotaFailOpen().Value.AsString()
	default:
		g.quotaReason = "quota_unknown"
		g.quotaStatus = GuardUnknown
	}
	h.Meter().AllowedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attr...)}...)
	return g
}

func (h *Handler) NewGuardError(err error) *Guard {
	return &Guard{
		Success:      GuardSuccess,
		Failure:      GuardFailure,
		Unknown:      GuardUnknown,
		finalStatus:  GuardUnknown,
		err:          err,
		quotaMessage: "",
		quotaToken:   "",
		quotaReason:  "quota_unknown",
		quotaStatus:  GuardUnknown,
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

func (h *Handler) ReasonKey(reason string) attribute.KeyValue {
	return reasonKey.String(reason)
}

func (h *Handler) ReasonFailOpen() attribute.KeyValue {
	return reasonKey.String("fail_open")
}

func (h *Handler) ReasonQuota() attribute.KeyValue {
	return reasonKey.String("quota")
}

func (h *Handler) ReasonQuotaCheckDisabled() attribute.KeyValue {
	return reasonKey.String("quota_check_disabled")
}

func (h *Handler) ReasonQuotaFailOpen() attribute.KeyValue {
	return reasonKey.String("quota_fail_open")
}

func (h *Handler) ReasonQuotaInvalidToken() attribute.KeyValue {
	return reasonKey.String("quota_invalid_token")
}

func (h *Handler) ReasonToken() attribute.KeyValue {
	return reasonKey.String("token")
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
