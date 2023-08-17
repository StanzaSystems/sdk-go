package handlers

import (
	"context"
	"time"

	"github.com/StanzaSystems/sdk-go/hub"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	GuardSuccess = iota
	GuardFailure
	GuardUnknown
)

type Guard struct {
	ctx   context.Context
	start time.Time
	meter *Meter
	attr  []attribute.KeyValue

	Success     int
	Failure     int
	Unknown     int
	finalStatus int
	err         error

	quotaToken   string
	quotaStatus  int
	quotaMessage string
	quotaReason  string
}

func (g *Guard) Allowed() bool {
	return g.quotaStatus != hub.CheckQuotaBlocked
}

func (g *Guard) Blocked() bool {
	return g.quotaStatus == hub.CheckQuotaBlocked
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

func (g *Guard) End(status int) {
	g.meter.AllowedDuration.Record(g.ctx,
		float64(time.Since(g.start).Microseconds())/1000,
		[]metric.RecordOption{metric.WithAttributes(g.attr...)}...)
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
