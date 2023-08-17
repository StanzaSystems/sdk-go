package httphandler

import (
	"context"
	"net/http"

	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/hub"
	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/StanzaSystems/sdk-go/otel"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

type InboundHandler struct {
	*handlers.InboundHandler
}

// NewInboundHandler returns a new InboundHandler
func NewInboundHandler() (*InboundHandler, error) {
	h, err := handlers.NewInboundHandler()
	if err != nil {
		return nil, err
	}
	return &InboundHandler{h}, nil
}

func (h *InboundHandler) VerifyServingCapacity(r *http.Request, route string, guard string) (context.Context, int) {
	ctx := h.Propagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	// Start HTTP server trace
	if route == "" {
		route = r.URL.Path
	}
	opts := []trace.SpanStartOption{
		trace.WithAttributes(httpconv.ServerRequest("", r)...),
		trace.WithSpanKind(trace.SpanKindServer),
	}
	ctx, span := h.Tracer().Start(ctx, route, opts...)
	defer span.End()

	// Clone TokenLeaseRequest (so we can add Feature and/or PriorityBoost)
	tlr := proto.Clone(h.TokenLeaseRequest(guard)).(*hubv1.GetTokenLeaseRequest)

	// Inspect Baggage and Headers for Feature and PriorityBoost, propagate through context if found
	ctx, tlr.Selector.FeatureName = otel.GetFeature(ctx, tlr.Selector.GetFeatureName())
	ctx, tlr.PriorityBoost = otel.GetPriorityBoost(ctx, tlr.GetPriorityBoost())

	// Add Guard and Feature to OTEL attributes
	attr := append(h.Attributes(),
		h.GuardKey(tlr.Selector.GetGuardName()),
		h.FeatureKey(tlr.Selector.GetFeatureName()),
	)

	if h.SentinelEnabled() {
		e, b := api.Entry(guard, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
		if b != nil {
			logging.Debug("Stanza blocked",
				"guard", guard,
				"sentinel.block_msg", b.BlockMsg(),
				"sentinel.block_type", b.BlockType().String(),
				"sentinel.block_value", b.TriggeredValue(),
				"sentinel.block_rule", b.TriggeredRule().String(),
			)
			attrWithReason := append(attr, h.ReasonKey(b.BlockType().String()))
			span.AddEvent("Stanza blocked", trace.WithAttributes(attrWithReason...))
			h.Meter().BlockedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
			return ctx, http.StatusTooManyRequests
		}
		e.Exit() // cleanly exit the Sentinel Entry
	}

	status := hub.ValidateTokens(ctx, guard, r.Header.Values("x-stanza-token"))
	if status == hub.ValidateTokensInvalid {
		attrWithReason := append(attr, h.ReasonQuotaInvalidToken())
		span.AddEvent("Stanza blocked", trace.WithAttributes(attrWithReason...))
		h.Meter().BlockedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
		return ctx, http.StatusTooManyRequests
	} else {
		attrWithReason := attr
		if status == hub.ValidateTokensFailOpen {
			attrWithReason = append(attrWithReason, h.ReasonFailOpen())
		}
		span.AddEvent("Stanza allowed", trace.WithAttributes(attrWithReason...))
		h.Meter().AllowedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
	}

	status, _ = hub.CheckQuota(ctx, tlr)
	if status == hub.CheckQuotaBlocked {
		attrWithReason := append(attr, h.ReasonQuota())
		span.AddEvent("Stanza blocked", trace.WithAttributes(attrWithReason...))
		h.Meter().BlockedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
		return ctx, http.StatusTooManyRequests
	} else {
		attrWithReason := attr
		if status == hub.CheckQuotaFailOpen {
			attrWithReason = append(attrWithReason, h.ReasonFailOpen())
		}
		if status == hub.CheckQuotaAllowed {
			attrWithReason = append(attrWithReason, h.ReasonQuota())
		}
		span.AddEvent("Stanza allowed", trace.WithAttributes(attrWithReason...))
		h.Meter().AllowedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
	}
	return ctx, http.StatusOK // return success
}
