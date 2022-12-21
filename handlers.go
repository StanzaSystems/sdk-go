package stanza

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
	instrumentationName = "github.com/StanzaSystems/sdk-go"

	httpServerAllowedCount   = "stanza.http.server.allowed"
	httpServerBlockedCount   = "stanza.http.server.blocked"
	httpServerTotalCount     = "stanza.http.server.total"
	httpServerDuration       = "stanza.http.server.duration"
	httpServerRequestSize    = "stanza.http.server.request.size"
	httpServerResponseSize   = "stanza.http.server.response.size"
	httpServerActiveRequests = "stanza.http.server.active"
)

type httpInboundMeters struct {
	resource       string
	allowedCount   syncint64.Counter
	blockedCount   syncint64.Counter
	totalCount     syncint64.Counter
	Duration       syncfloat64.Histogram
	RequestSize    syncint64.Histogram
	ResponseSize   syncint64.Histogram
	ActiveRequests syncint64.UpDownCounter
}

func InitHttpInboundHandler(res string) (*httpInboundMeters, error) {
	meter := global.MeterProvider().Meter(
		instrumentationName,                        // TODO: should this be a customer "DSN" of some form?
		metric.WithInstrumentationVersion("0.0.1"), // TODO: stanza sdk-go version goes here
	)

	var err error
	him := &httpInboundMeters{resource: res}

	// sentinel meters
	him.allowedCount, err = meter.SyncInt64().Counter(
		httpServerAllowedCount,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of inbound HTTP requests that were allowed"))
	if err != nil {
		return nil, err
	}
	him.blockedCount, err = meter.SyncInt64().Counter(
		httpServerBlockedCount,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of inbound HTTP requests that were backpressured"))
	if err != nil {
		return nil, err
	}
	him.totalCount, err = meter.SyncInt64().Counter(
		httpServerTotalCount,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of inbound HTTP requests that were checked"))
	if err != nil {
		return nil, err
	}

	// generic HTTP server meters
	him.Duration, err = meter.SyncFloat64().Histogram(
		httpServerDuration,
		instrument.WithUnit(unit.Milliseconds),
		instrument.WithDescription("measures the duration inbound HTTP requests"))
	if err != nil {
		return nil, err
	}
	him.RequestSize, err = meter.SyncInt64().Histogram(
		httpServerRequestSize,
		instrument.WithUnit(unit.Bytes),
		instrument.WithDescription("measures the size of HTTP request messages"))
	if err != nil {
		return nil, err
	}
	him.ResponseSize, err = meter.SyncInt64().Histogram(
		httpServerResponseSize,
		instrument.WithUnit(unit.Bytes),
		instrument.WithDescription("measures the size of HTTP response messages"))
	if err != nil {
		return nil, err
	}
	him.ActiveRequests, err = meter.SyncInt64().UpDownCounter(
		httpServerActiveRequests,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of concurrent HTTP requests in-flight"))
	if err != nil {
		return nil, err
	}

	return him, nil
}

func HttpInboundHandler(ctx context.Context, him *httpInboundMeters, req *http.Request) int {
	// Wrap OTEL( (https://github.com/gofiber/contrib/tree/main/otelfiber))

	e, b := api.Entry(him.resource, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
	if b != nil {
		him.blockedCount.Add(ctx, 1)
		// blocked count by sentinel block type?
		him.totalCount.Add(ctx, 1)
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
		him.allowedCount.Add(ctx, 1)
		him.totalCount.Add(ctx, 1)
		// latency percentiles?

		// Be sure the entry is exited finally.
		e.Exit()
		return http.StatusOK
	}
}
