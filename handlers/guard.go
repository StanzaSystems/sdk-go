package handlers

import (
	"context"
	"fmt"
	"time"

	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/global"
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
	tlr   *hubv1.GetTokenLeaseRequest
	meter *global.StanzaMeter
	span  trace.Span
	attr  []attribute.KeyValue
	err   error

	Success int
	Failure int
	Unknown int

	configStatus hubv1.Config
	config       *hubv1.GuardConfig

	localStatus hubv1.Local
	localBlock  *base.BlockError

	tokenStatus hubv1.Token

	quotaStatus hubv1.Quota
	quotaToken  string
}

func (g *Guard) Allowed() bool {
	if (g.configStatus == hubv1.Config_CONFIG_CACHED_OK ||
		g.configStatus == hubv1.Config_CONFIG_FETCHED_OK) &&
		g.localStatus != hubv1.Local_LOCAL_BLOCKED &&
		g.quotaStatus != hubv1.Quota_QUOTA_BLOCKED &&
		g.tokenStatus != hubv1.Token_TOKEN_NOT_VALID {
		return true
	}
	return false
}

func (g *Guard) Blocked() bool {
	if g.configStatus == hubv1.Config_CONFIG_UNSPECIFIED ||
		g.configStatus == hubv1.Config_CONFIG_FETCH_ERROR ||
		g.configStatus == hubv1.Config_CONFIG_FETCH_TIMEOUT ||
		g.configStatus == hubv1.Config_CONFIG_NOT_FOUND ||
		g.localStatus == hubv1.Local_LOCAL_BLOCKED ||
		g.quotaStatus == hubv1.Quota_QUOTA_BLOCKED ||
		g.tokenStatus == hubv1.Token_TOKEN_NOT_VALID {
		return true
	}
	return false
}

func (g *Guard) BlockMessage() string {
	if g.localStatus == hubv1.Local_LOCAL_BLOCKED {
		return g.localBlock.BlockMsg()
	}
	if g.tokenStatus == hubv1.Token_TOKEN_NOT_VALID {
		return "Invalid or expired X-Stanza-Token."
	}
	if g.quotaStatus == hubv1.Quota_QUOTA_BLOCKED {
		return "Stanza quota exhausted. Please try again later."
	}
	return ""
}

func (g *Guard) BlockReason() string {
	if g.localStatus == hubv1.Local_LOCAL_BLOCKED {
		return hubv1.Local_name[int32(hubv1.Local_LOCAL_BLOCKED)]
	}
	if g.tokenStatus == hubv1.Token_TOKEN_NOT_VALID {
		return hubv1.Token_name[int32(hubv1.Token_TOKEN_NOT_VALID)]
	}
	if g.quotaStatus == hubv1.Quota_QUOTA_BLOCKED {
		return hubv1.Quota_name[int32(hubv1.Quota_QUOTA_BLOCKED)]
	}
	return ""
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
}

func (g *Guard) getGuardConfig(ctx context.Context, name string) (hubv1.Config, error) {
	gc, err := global.GetGuardConfig(ctx, name)
	if err != nil {
		logging.Error(err)
		g.failopen(ctx, err)
		g.configStatus = hubv1.Config_CONFIG_FETCH_ERROR
		return g.configStatus, err
	} else {
		g.config = gc
		g.configStatus = hubv1.Config_CONFIG_CACHED_OK
		return g.configStatus, nil
	}
}

func (g *Guard) checkLocal(ctx context.Context, name string, enabled bool) error {
	if !enabled {
		g.localStatus = hubv1.Local_LOCAL_EVAL_DISABLED
	} else {
		e, b := api.Entry(name, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
		if b != nil {
			g.localBlock = b
			logging.Debug("Sentinel blocked",
				"guard", name,
				"sentinel.block_msg", b.BlockMsg(),
				"sentinel.block_type", b.BlockType().String(),
				"sentinel.block_value", b.TriggeredValue(),
				"sentinel.block_rule", b.TriggeredRule().String(),
			)
			g.blocked(ctx)
			g.localStatus = hubv1.Local_LOCAL_BLOCKED
		} else {
			e.Exit() // cleanly exit the Sentinel Entry
			g.localStatus = hubv1.Local_LOCAL_ALLOWED
		}
	}
	return nil
}

func (g *Guard) checkToken(ctx context.Context, name string, tokens []string, enabled bool) error {
	if !enabled {
		g.tokenStatus = hubv1.Token_TOKEN_EVAL_DISABLED
	} else {
		g.tokenStatus, g.err = hub.ValidateTokens(ctx, name, tokens)
		if g.err != nil {
			g.failopen(ctx, g.err)
		}
		if g.tokenStatus == hubv1.Token_TOKEN_NOT_VALID {
			g.blocked(ctx)
		}
	}
	return g.err
}

func (g *Guard) checkQuota(ctx context.Context, tlr *hubv1.GetTokenLeaseRequest, enabled bool) error {
	g.tlr = tlr
	if !enabled {
		g.quotaStatus = hubv1.Quota_QUOTA_EVAL_DISABLED
	} else {
		g.quotaStatus, g.quotaToken, g.err = hub.CheckQuota(ctx, tlr)
		if g.err != nil {
			g.failopen(ctx, g.err)
		}
		if g.quotaStatus == hubv1.Quota_QUOTA_BLOCKED {
			g.blocked(ctx)
		}
	}
	return g.err
}

func (g *Guard) allowed(ctx context.Context) {
	g.meter.AllowedCount.Add(ctx, 1, g.metricAttr()...)
	g.span.AddEvent("Stanza allowed", g.traceAttr(nil))
	logging.Debug("Stanza allowed", g.logReasons(nil)...)
}

func (g *Guard) blocked(ctx context.Context) {
	g.meter.BlockedCount.Add(ctx, 1, g.metricAttr()...)
	g.span.AddEvent("Stanza blocked", g.traceAttr(nil))
	logging.Debug("Stanza blocked", g.logReasons(nil)...)
}

func (g *Guard) failopen(ctx context.Context, err error) {
	g.meter.FailOpenCount.Add(ctx, 1, g.metricAttr()...)
	g.span.AddEvent("Stanza failed open", g.traceAttr(err))
	logging.Debug("Stanza failed open", g.logReasons(err)...)
}

func (g *Guard) metricAttr() []metric.AddOption {
	return []metric.AddOption{metric.WithAttributes(g.reasons()...)}
}

func (g *Guard) traceAttr(err error) trace.SpanStartEventOption {
	var resp []attribute.KeyValue
	if err != nil {
		resp = append(resp, errorKey.String(err.Error()))
	}
	resp = append(resp, g.reasons()...)
	return trace.WithAttributes(resp...)
}

func (g *Guard) reasons() []attribute.KeyValue {
	kvs := g.attr
	kvs = append(kvs, configReasonKey.String(hubv1.Config_name[int32(g.configStatus)]))
	kvs = append(kvs, localReasonKey.String(hubv1.Local_name[int32(g.localStatus)]))
	kvs = append(kvs, tokenReasonKey.String(hubv1.Token_name[int32(g.tokenStatus)]))
	kvs = append(kvs, quotaReasonKey.String(hubv1.Quota_name[int32(g.quotaStatus)]))
	return kvs
}

func (g *Guard) logReasons(err error) []interface{} {
	resp := make([]interface{}, 0, 10)
	if err != nil {
		resp = append(resp, "error", err.Error())
	}
	if g.tlr != nil {
		if g.tlr.Selector != nil {
			resp = append(resp, "guard", g.tlr.GetSelector().GetGuardName())
			if g.tlr.GetSelector().FeatureName != nil {
				resp = append(resp, "feature", g.tlr.GetSelector().GetFeatureName())
			}
		}
		if g.tlr.PriorityBoost != nil {
			resp = append(resp, "priority_boost", fmt.Sprintf("%d", g.tlr.GetPriorityBoost()))
		}
		if g.tlr.DefaultWeight != nil {
			resp = append(resp, "default_weight", fmt.Sprintf("%.2f", g.tlr.GetDefaultWeight()))
		}
	}
	return append(resp,
		localReason, hubv1.Local_name[int32(g.localStatus)],
		tokenReason, hubv1.Token_name[int32(g.tokenStatus)],
		quotaReason, hubv1.Quota_name[int32(g.quotaStatus)],
	)
}
