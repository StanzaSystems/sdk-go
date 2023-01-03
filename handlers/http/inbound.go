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
		instrumentationName,                        // TODO: should this be a customer "DSN" of some form?
		metric.WithInstrumentationVersion("0.0.1"), // TODO: stanza sdk-go version goes here
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

func InboundHandler(ctx context.Context, im *InboundMeters, req *http.Request) (context.Context, int) {
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(req.Header))
	opts := []trace.SpanStartOption{
		trace.WithAttributes(
			semconv.HTTPServerNameKey.String("TODO"),
			semconv.HTTPMethodKey.String(req.Method),
			semconv.HTTPURLKey.String(req.URL.String()),
			// semconv.NetHostIPKey.String(utils.CopyString(c.IP())),
			// semconv.NetHostNameKey.String(utils.CopyString(c.Hostname())),
			// semconv.HTTPUserAgentKey.String(string(utils.CopyBytes(c.Request().Header.UserAgent()))),
			semconv.HTTPRequestContentLengthKey.Int64(req.ContentLength),
			semconv.HTTPSchemeKey.String(string(req.Proto)),
			semconv.NetTransportTCP),
		trace.WithSpanKind(trace.SpanKindServer),
	}
	// if len(c.IPs()) > 0 {
	// 	opts = append(opts, trace.WithAttributes(semconv.HTTPClientIPKey.String(utils.CopyString(c.IPs()[0]))))
	// }

	tracer := otel.GetTracerProvider().Tracer("InstrumentationName")
	ctx, span := tracer.Start(ctx, "TodoNameSpan", opts...)
	defer span.End()

	e, b := api.Entry(im.resource, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
	if b != nil {
		// TODO: allow sentinel "customize block fallback" to override this 429 default
		sc := http.StatusTooManyRequests

		byTypeAttrs := append(im.Attributes, attribute.String("sentinel.block.type", b.BlockType().String()))
		im.totalCount.Add(ctx, 1, im.Attributes...)
		im.blockedCount.Add(ctx, 1, im.Attributes...)
		im.blockedCountByType.Add(ctx, 1, byTypeAttrs...)

		// TODO: add additional sentinel specific info to span?
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(sc)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(sc, trace.SpanKindServer)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)

		logging.Error(nil, "Stanza blocked",
			"BlockMessage", b.BlockMsg(),
			"BlockType", b.BlockType().String(),
			"BlockValue", b.TriggeredValue(),
		)
		logging.Debug("BlockRule", b.TriggeredRule().String())
		return ctx, sc
	
	} else {
		sc := http.StatusOK

		im.totalCount.Add(ctx, 1, im.Attributes...)
		im.allowedCount.Add(ctx, 1, im.Attributes...)

		// TODO: add additional sentinel specific info to span?
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(sc)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(sc, trace.SpanKindServer)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)

		e.Exit()       // cleanly exit the Sentinel Entry
		return ctx, sc // return success
	}
}
