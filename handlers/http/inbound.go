package http

import (
	"context"
	"net/http"

	sg "github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	httpServerAllowedCount       = "stanza.http.server.sentinel.allowed"
	httpServerBlockedCount       = "stanza.http.server.sentinel.blocked"
	httpServerBlockedCountByType = "stanza.http.server.sentinel.blocked.by"
	httpServerTotalCount         = "stanza.http.server.sentinel.total"
	httpServerDuration           = "stanza.http.server.duration"
	httpServerRequestSize        = "stanza.http.server.request.size"
	httpServerResponseSize       = "stanza.http.server.response.size"
	httpServerActiveRequests     = "stanza.http.server.active"
)

type InboundMeters struct {
	resource           string
	allowedCount       syncint64.Counter
	blockedCount       syncint64.Counter
	blockedCountByType syncint64.Counter
	totalCount         syncint64.Counter

	Attributes     []attribute.KeyValue
	Duration       syncfloat64.Histogram
	RequestSize    syncint64.Histogram
	ResponseSize   syncint64.Histogram
	ActiveRequests syncint64.UpDownCounter
}

func InitInboundMeters(res string) (InboundMeters, error) {
	meter := global.MeterProvider().Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(instrumentationVersion),
	)

	var err error
	im := InboundMeters{
		resource:   res,
		Attributes: []attribute.KeyValue{attribute.String("sentinel.resource", res)},
	}

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
	im.blockedCountByType, err = meter.SyncInt64().Counter(
		httpServerBlockedCountByType,
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

func InboundHandler(ctx context.Context, route string, im *InboundMeters, req *http.Request) (context.Context, int) {
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(req.Header))

	// generic HTTP server trace
	opts := []trace.SpanStartOption{
		trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", req)...),
		trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(sg.Name(), route, req)...),
		trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(req)...),
		trace.WithSpanKind(trace.SpanKindServer),
	}
	tracer := otel.GetTracerProvider().Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(instrumentationVersion),
	)
	ctx, span := tracer.Start(ctx, route, opts...)
	defer span.End()

	e, b := api.Entry(im.resource, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
	if b != nil {
		// TODO: allow sentinel "customize block fallback" to override this 429 default
		sc := http.StatusTooManyRequests // default `429 Too Many Request`
		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(sc)...)

		byTypeAttrs := append(im.Attributes, attribute.String("sentinel.block.type", b.BlockType().String()))
		im.totalCount.Add(ctx, 1, im.Attributes...)
		im.blockedCount.Add(ctx, 1, im.Attributes...)
		im.blockedCountByType.Add(ctx, 1, byTypeAttrs...)

		// TODO: add additional sentinel specific info to span
		// (at least im.resource, b.BlockMessage, b.BlockType, and b.BlockValue)

		logging.Error(nil, "Stanza blocked",
			"BlockMessage", b.BlockMsg(),
			"BlockType", b.BlockType().String(),
			"BlockValue", b.TriggeredValue(),
		)
		logging.Debug("BlockRule", b.TriggeredRule().String())
		return ctx, sc
	
	} else {
		sc := http.StatusOK
		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(sc)...)

		im.totalCount.Add(ctx, 1, im.Attributes...)
		im.allowedCount.Add(ctx, 1, im.Attributes...)

		// TODO: add additional sentinel specific info to span?

		e.Exit()       // cleanly exit the Sentinel Entry
		return ctx, sc // return success
	}
}
