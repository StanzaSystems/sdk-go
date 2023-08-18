package handlers

import (
	"context"
	"time"

	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/hub"
	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	GuardUnknown = iota
	GuardSuccess
	GuardFailure
)

type Guard struct {
	ctx   context.Context
	start time.Time
	meter *Meter
	span  trace.Span
	attr  []attribute.KeyValue
	err   error

	Success     int
	Failure     int
	Unknown     int
	finalStatus int

	quotaToken   string
	quotaStatus  int
	quotaMessage string
	quotaReason  string

	sentinelBlock *base.BlockError
}

func (g *Guard) Allowed() bool {
	if g.sentinelBlock == nil && g.quotaStatus != hub.CheckQuotaBlocked {
		return true
	}
	return false
}

func (g *Guard) Blocked() bool {
	if g.sentinelBlock != nil || g.quotaStatus == hub.CheckQuotaBlocked {
		return true
	}
	return false
}

func (g *Guard) BlockMessage() string {
	return g.quotaMessage
}

func (g *Guard) BlockReason() string {
	return g.quotaReason
}

func (g *Guard) Token() string {
	return g.quotaToken
}

func (g *Guard) Error() error {
	return g.err
}

func (g *Guard) Context() context.Context {
	return g.ctx
}

func (g *Guard) End(status int) {
	if !g.start.IsZero() {
		g.meter.AllowedDuration.Record(g.ctx,
			float64(time.Since(g.start).Microseconds())/1000,
			[]metric.RecordOption{metric.WithAttributes(g.attr...)}...)
	}
	if status == g.Success {
		g.meter.AllowedSuccessCount.Add(g.ctx, 1, []metric.AddOption{metric.WithAttributes(g.attr...)}...)
	}
	if status == g.Failure {
		g.meter.AllowedFailureCount.Add(g.ctx, 1, []metric.AddOption{metric.WithAttributes(g.attr...)}...)
	}
	if status == g.Unknown {
		g.meter.AllowedUnknownCount.Add(g.ctx, 1, []metric.AddOption{metric.WithAttributes(g.attr...)}...)
	}
	g.finalStatus = status
}

func (g *Guard) checkSentinel(name string) {
	attrWithReason := append(g.attr, reason(ReasonSentinel))
	e, b := api.Entry(name, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
	if b != nil {
		g.sentinelBlock = b
		logging.Debug("Stanza blocked",
			"guard", name,
			"sentinel.block_msg", b.BlockMsg(),
			"sentinel.block_type", b.BlockType().String(),
			"sentinel.block_value", b.TriggeredValue(),
			"sentinel.block_rule", b.TriggeredRule().String(),
		)
		g.span.AddEvent("Stanza blocked", trace.WithAttributes(attrWithReason...))
		g.meter.BlockedCount.Add(g.ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
	} else {
		g.span.AddEvent("Stanza allowed", trace.WithAttributes(attrWithReason...))
		g.meter.AllowedCount.Add(g.ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
	}
	e.Exit() // cleanly exit the Sentinel Entry
}

func (g *Guard) checkQuota(ctx context.Context, tlr *hubv1.GetTokenLeaseRequest) {
	g.quotaStatus, g.quotaToken = hub.CheckQuota(ctx, tlr)

	attrWithReason := g.attr
	switch g.quotaStatus {
	case hub.CheckQuotaBlocked:
		attrWithReason = append(attrWithReason, reason(ReasonQuota))
		g.quotaReason = reason(ReasonQuota).Value.AsString()
		g.meter.BlockedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
		g.span.AddEvent("Stanza blocked", trace.WithAttributes(attrWithReason...))
		logging.Debug("Stanza blocked",
			"guard", tlr.GetSelector().GetGuardName(),
			"feature", tlr.GetSelector().GetFeatureName(),
			"default_weight", tlr.DefaultWeight,
			"priority_boost", tlr.PriorityBoost,
			"reason", g.quotaReason,
		)
		return
	case hub.CheckQuotaAllowed:
		attrWithReason = append(attrWithReason, reason(ReasonQuota))
		g.quotaReason = reason(ReasonQuota).Value.AsString()
	case hub.CheckQuotaSkipped:
		attrWithReason = append(attrWithReason, reason(ReasonQuotaCheckDisabled))
		g.quotaReason = reason(ReasonQuotaCheckDisabled).Value.AsString()
	case hub.CheckQuotaFailOpen:
		attrWithReason = append(attrWithReason, reason(ReasonQuotaFailOpen))
		g.quotaReason = reason(ReasonQuotaFailOpen).Value.AsString()
	default:
		attrWithReason = append(attrWithReason, reason(ReasonQuotaUnknown))
		g.quotaReason = reason(ReasonQuotaUnknown).Value.AsString()
		g.quotaStatus = GuardUnknown
	}
	g.span.AddEvent("Stanza allowed", trace.WithAttributes(attrWithReason...))
	g.meter.AllowedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
}

func (g *Guard) checkToken(ctx context.Context, name string, tokens []string) {
	status := hub.ValidateTokens(ctx, name, tokens)

	attrWithReason := g.attr
	switch status {
	case hub.ValidateTokensInvalid:
		attrWithReason := append(attrWithReason, reason(ReasonQuotaToken))
		g.quotaReason = reason(ReasonQuotaToken).Value.AsString()
		g.meter.BlockedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
		g.span.AddEvent("Stanza blocked", trace.WithAttributes(attrWithReason...))
		logging.Debug("Stanza blocked",
			"guard", name,
			"reason", g.quotaReason,
		)
		return
	case hub.ValidateTokensValid:
		attrWithReason = append(attrWithReason, reason(ReasonQuotaToken))
		g.quotaReason = reason(ReasonQuotaToken).Value.AsString()
	case hub.ValidateTokensFailOpen:
		attrWithReason = append(attrWithReason, reason(ReasonQuotaFailOpen))
		g.quotaReason = reason(ReasonQuotaFailOpen).Value.AsString()
	case hub.ValidateTokensSkipped:
		attrWithReason = append(attrWithReason, reason(ReasonQuotaCheckDisabled))
		g.quotaReason = reason(ReasonQuotaCheckDisabled).Value.AsString()
	}
	g.span.AddEvent("Stanza allowed", trace.WithAttributes(attrWithReason...))
	g.meter.AllowedCount.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(attrWithReason...)}...)
}
