package http

import (
	"context"
	"net/http"

	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
)

const (
	httpServerAllowedCount   = "stanza.http.server.sentinel.allowed"
	httpServerBlockedCount   = "stanza.http.server.sentinel.blocked"
	httpServerTotalCount     = "stanza.http.server.sentinel.total"
	httpServerDuration       = "stanza.http.server.duration"
	httpServerRequestSize    = "stanza.http.server.request.size"
	httpServerResponseSize   = "stanza.http.server.response.size"
	httpServerActiveRequests = "stanza.http.server.active"
)

type InboundMeters struct {
	resource       string
	allowedCount   syncint64.Counter
	blockedCount   syncint64.Counter
	totalCount     syncint64.Counter
	Duration       syncfloat64.Histogram
	RequestSize    syncint64.Histogram
	ResponseSize   syncint64.Histogram
	ActiveRequests syncint64.UpDownCounter
}

func InitInboundMeters(res string) (InboundMeters, error) {
	meter := global.MeterProvider().Meter(
		instrumentationName,                        // TODO: should this be a customer "DSN" of some form?
		metric.WithInstrumentationVersion("0.0.1"), // TODO: stanza sdk-go version goes here
	)

	var err error
	im := InboundMeters{resource: res}

	// sentinel meters
	im.allowedCount, err = meter.SyncInt64().Counter(
		httpServerAllowedCount,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of inbound HTTP requests that were allowed"))
	if err != nil {
		return im, err
	}
	im.blockedCount, err = meter.SyncInt64().Counter(
		httpServerBlockedCount,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of inbound HTTP requests that were backpressured"))
	if err != nil {
		return im, err
	}
	im.totalCount, err = meter.SyncInt64().Counter(
		httpServerTotalCount,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of inbound HTTP requests that were checked"))
	if err != nil {
		return im, err
	}

	// generic HTTP server meters
	im.Duration, err = meter.SyncFloat64().Histogram(
		httpServerDuration,
		instrument.WithUnit(unit.Milliseconds),
		instrument.WithDescription("measures the duration inbound HTTP requests"))
	if err != nil {
		return im, err
	}
	im.RequestSize, err = meter.SyncInt64().Histogram(
		httpServerRequestSize,
		instrument.WithUnit(unit.Bytes),
		instrument.WithDescription("measures the size of HTTP request messages"))
	if err != nil {
		return im, err
	}
	im.ResponseSize, err = meter.SyncInt64().Histogram(
		httpServerResponseSize,
		instrument.WithUnit(unit.Bytes),
		instrument.WithDescription("measures the size of HTTP response messages"))
	if err != nil {
		return im, err
	}
	im.ActiveRequests, err = meter.SyncInt64().UpDownCounter(
		httpServerActiveRequests,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of concurrent HTTP requests in-flight"))
	if err != nil {
		return im, err
	}

	return im, nil
}

func InboundHandler(ctx context.Context, im *InboundMeters, req *http.Request) int {
	e, b := api.Entry(im.resource, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
	if b != nil {
		// TODO: all of these need to be "sentinel resource" tagged!!
		im.blockedCount.Add(ctx, 1)
		// blocked count by sentinel block type?
		im.totalCount.Add(ctx, 1)
		// latency percentiles?

		// what do we want from that http request?
		// I think potentially a lot for our trace/span...
		// not sure about metrics though -- maybe some "path" based counts?

		logging.Error(nil, "Stanza blocked",
			"BlockMessage", b.BlockMsg(),
			"BlockType", b.BlockType().String(),
			"BlockValue", b.TriggeredValue(),
		)
		logging.Debug("BlockRule", b.TriggeredRule().String())

		// TODO: allow sentinel "customize block fallback" to override this 429 default
		return http.StatusTooManyRequests
	} else {
		im.allowedCount.Add(ctx, 1)
		im.totalCount.Add(ctx, 1)
		// latency percentiles?

		e.Exit()             // cleanly exit the sentinel entry
		return http.StatusOK // return success
	}
}
