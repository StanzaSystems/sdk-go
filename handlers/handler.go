package handlers

import (
	"context"
	"time"

	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
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
	attr          []attribute.KeyValue
}

func NewHandler(gn string, fn *string, pb *int32, dw *float32) (*Handler, error) {
	return &Handler{
		guardName:     gn,
		featureName:   fn,
		priorityBoost: pb,
		defaultWeight: dw,
		attr: []attribute.KeyValue{
			clientIdKey.String(global.GetClientID()),
			environmentKey.String(global.GetServiceEnvironment()),
			serviceKey.String(global.GetServiceName()),
		},
	}, nil
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

	// Config State check
	_, err := g.getGuardConfig(ctx, h.guardName)
	if err != nil || g.config == nil {
		return g
	}

	// Local (Sentinel) check
	err = g.checkLocal(ctx, h.guardName, h.SentinelEnabled())
	if err != nil || g.localStatus == hubv1.Local_LOCAL_BLOCKED {
		return g
	}

	// Ingress token check
	err = g.checkToken(ctx, h.guardName, tokens, g.config.ValidateIngressTokens)
	if err != nil || g.tokenStatus == hubv1.Token_TOKEN_NOT_VALID {
		return g
	}

	// Quota check
	err = g.checkQuota(ctx, tlr, g.config.CheckQuota)
	if err != nil || g.quotaStatus == hubv1.Quota_QUOTA_BLOCKED {
		return g
	}

	g.allowed(ctx)
	return g
}

func (h *Handler) NewGuard(ctx context.Context, span trace.Span, attr []attribute.KeyValue, err error) *Guard {
	attr = append(attr, customerIdKey.String(global.GetCustomerID()))
	return &Guard{
		ctx:   ctx,
		start: time.Time{},
		tlr:   &hubv1.GetTokenLeaseRequest{Selector: &hubv1.GuardFeatureSelector{GuardName: h.guardName}},
		meter: global.GetStanzaMeter(),
		span:  span,
		attr:  append(h.attr, attr...),
		err:   err,

		Success: GuardSuccess,
		Failure: GuardFailure,
		Unknown: GuardUnknown,

		configStatus: hubv1.Config_CONFIG_NOT_FOUND,
		config:       nil,
		localStatus:  hubv1.Local_LOCAL_NOT_EVAL,
		localBlock:   nil,
		tokenStatus:  hubv1.Token_TOKEN_NOT_EVAL,
		quotaStatus:  hubv1.Quota_QUOTA_NOT_EVAL,
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
	return *global.GetStanzaTracer()
}

func (h *Handler) Propagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

func (h *Handler) FailOpen(ctx context.Context) {
	if m := global.GetStanzaMeter(); m != nil {
		m.FailOpenCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(h.attr...)}...)
	}
}

// Global Helper Functions //
func (h *Handler) OTELEnabled() bool {
	return global.OtelEnabled()
}

func (h *Handler) SentinelEnabled() bool {
	return global.SentinelEnabled()
}
