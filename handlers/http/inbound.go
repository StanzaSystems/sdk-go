package http

import (
	"context"
	"net/http"

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
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
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
	Attributes []attribute.KeyValue

	AllowedCount       syncint64.Counter
	BlockedCount       syncint64.Counter
	BlockedCountByType syncint64.Counter
	TotalCount         syncint64.Counter
	Duration           syncfloat64.Histogram
	RequestSize        syncint64.Histogram
	ResponseSize       syncint64.Histogram
	ActiveRequests     syncint64.UpDownCounter
}

func InitInboundMeters(decorator string) (InboundMeters, error) {
	meter := global.MeterProvider().Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(instrumentationVersion),
	)

	var err error
	im := InboundMeters{
		Attributes: []attribute.KeyValue{attribute.String("stanza.decorator", decorator)},
	}

	// sentinel meters
	im.AllowedCount, err = meter.SyncInt64().Counter(
		httpServerAllowedCount,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of inbound HTTP requests that were allowed"))
	if err != nil {
		return im, err
	}
	im.BlockedCount, err = meter.SyncInt64().Counter(
		httpServerBlockedCount,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of inbound HTTP requests that were backpressured"))
	if err != nil {
		return im, err
	}
	im.BlockedCountByType, err = meter.SyncInt64().Counter(
		httpServerBlockedCountByType,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of inbound HTTP requests that were backpressured"))
	if err != nil {
		return im, err
	}
	im.TotalCount, err = meter.SyncInt64().Counter(
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

func InboundHandler(ctx context.Context, name, decorator, route string, im *InboundMeters, req *http.Request) (context.Context, int) {
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(req.Header))

	// generic HTTP server trace
	opts := []trace.SpanStartOption{
		trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", req)...),
		trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(name, route, req)...),
		trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(req)...),
		trace.WithSpanKind(trace.SpanKindServer),
	}
	tracer := otel.GetTracerProvider().Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(instrumentationVersion),
	)
	ctx, span := tracer.Start(ctx, route, opts...)
	defer span.End()

	e, b := api.Entry(decorator, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
	if b != nil {
		// TODO: allow sentinel "customize block fallback" to override this 429 default
		sc := http.StatusTooManyRequests // default `429 Too Many Request`
		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(sc)...)

		byTypeAttrs := append(im.Attributes, attribute.String("sentinel.block.type", b.BlockType().String()))
		im.TotalCount.Add(ctx, 1, im.Attributes...)
		im.BlockedCount.Add(ctx, 1, im.Attributes...)
		im.BlockedCountByType.Add(ctx, 1, byTypeAttrs...)

		// TODO: add additional sentinel specific info to span
		// (at least decorator, b.BlockMessage, b.BlockType, and b.BlockValue)

		logging.Error(nil, "stanza blocked",
			"BlockMessage", b.BlockMsg(),
			"BlockType", b.BlockType().String(),
			"BlockValue", b.TriggeredValue(),
		)
		logging.Debug("stanza blocked",
			"BlockRule", b.TriggeredRule().String(),
		)
		return ctx, sc

	} else {
		sc := http.StatusOK
		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(sc)...)

		im.TotalCount.Add(ctx, 1, im.Attributes...)
		im.AllowedCount.Add(ctx, 1, im.Attributes...)

		// TODO: add additional sentinel specific info to span?

		e.Exit()       // cleanly exit the Sentinel Entry
		return ctx, sc // return success
	}
}
