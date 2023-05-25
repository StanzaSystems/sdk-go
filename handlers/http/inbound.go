package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv/v1.17.0/httpconv"
	"go.opentelemetry.io/otel/trace"
)

const (
	httpServerDecorator          = "stanza.http.server.request.decorator"
	httpServerAllowedCount       = "stanza.http.server.sentinel.allowed"
	httpServerBlockedCount       = "stanza.http.server.sentinel.blocked"
	httpServerBlockedCountByType = "stanza.http.server.sentinel.blocked.by"
	httpServerBlockedMessage     = "stanza.http.server.sentinel.blocked.message"
	httpServerBlockedType        = "stanza.http.server.sentinel.blocked.type"
	httpServerBlockedValue       = "stanza.http.server.sentinel.blocked.value"
	httpServerBlockedRule        = "stanza.http.server.sentinel.blocked.rule"
	httpServerTotalCount         = "stanza.http.server.sentinel.total"
	httpServerDuration           = "stanza.http.server.duration"
	httpServerRequestSize        = "stanza.http.server.request.size"
	httpServerResponseSize       = "stanza.http.server.response.size"
	httpServerActiveRequests     = "stanza.http.server.active"
)

var (
	debugBaggageKey   = attribute.Key("hub.getstanza.io/StanzaDebug")
	decoratorKey      = attribute.Key(httpServerDecorator)
	blockedMessageKey = attribute.Key(httpServerBlockedMessage)
	blockedTypeKey    = attribute.Key(httpServerBlockedType)
	blockedRuleKey    = attribute.Key(httpServerBlockedRule)
)

type InboundMeters struct {
	Attributes []attribute.KeyValue

	AllowedCount       metric.Int64Counter
	BlockedCount       metric.Int64Counter
	BlockedCountByType metric.Int64Counter
	TotalCount         metric.Int64Counter
	Duration           metric.Float64Histogram
	RequestSize        metric.Int64Histogram
	ResponseSize       metric.Int64Histogram
	ActiveRequests     metric.Int64UpDownCounter
}

type InboundHandler struct {
	app         string
	decorator   string
	sentinel    bool
	propagators propagation.TextMapPropagator
	tracer      trace.Tracer
	meter       InboundMeters
}

// New returns a new InboundHandler
func NewInboundHandler(app, decorator string, sentinel bool) (*InboundHandler, error) {
	handler := &InboundHandler{
		app:         app,
		decorator:   decorator,
		sentinel:    sentinel,
		propagators: otel.GetTextMapPropagator(),
		tracer: otel.GetTracerProvider().Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(instrumentationVersion),
		),
	}
	meter := otel.Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(instrumentationVersion),
	)
	im := InboundMeters{
		Attributes: []attribute.KeyValue{decoratorKey.String(decorator)},
	}

	// sentinel meters
	var err error
	im.AllowedCount, err = meter.Int64Counter(
		httpServerAllowedCount,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of inbound HTTP requests that were allowed"))
	if err != nil {
		return handler, err
	}
	im.BlockedCount, err = meter.Int64Counter(
		httpServerBlockedCount,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of inbound HTTP requests that were backpressured"))
	if err != nil {
		return handler, err
	}
	im.BlockedCountByType, err = meter.Int64Counter(
		httpServerBlockedCountByType,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of inbound HTTP requests that were backpressured"))
	if err != nil {
		return handler, err
	}
	im.TotalCount, err = meter.Int64Counter(
		httpServerTotalCount,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of inbound HTTP requests that were checked"))
	if err != nil {
		return handler, err
	}

	// generic HTTP server meters
	im.Duration, err = meter.Float64Histogram(
		httpServerDuration,
		metric.WithUnit("ms"),
		metric.WithDescription("measures the duration inbound HTTP requests"))
	if err != nil {
		return handler, err
	}
	im.RequestSize, err = meter.Int64Histogram(
		httpServerRequestSize,
		metric.WithUnit("By"),
		metric.WithDescription("measures the size of HTTP request messages"))
	if err != nil {
		return handler, err
	}
	im.ResponseSize, err = meter.Int64Histogram(
		httpServerResponseSize,
		metric.WithUnit("By"),
		metric.WithDescription("measures the size of HTTP response messages"))
	if err != nil {
		return handler, err
	}
	im.ActiveRequests, err = meter.Int64UpDownCounter(
		httpServerActiveRequests,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of concurrent HTTP requests in-flight"))
	if err != nil {
		return handler, err
	}

	handler.meter = im
	return handler, nil
}

func (h *InboundHandler) Meter() *InboundMeters {
	return &h.meter
}

// func InboundHandler(ctx context.Context, name, decorator, route string, im *InboundMeters, req *http.Request) (context.Context, int) {
func (h *InboundHandler) VerifyServingCapacity(r *http.Request, route string) (context.Context, int) {
	ctx := h.propagators.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	m0, _ := baggage.NewMember(string(debugBaggageKey), "TRUE")
	b0, _ := baggage.New(m0)
	ctx = baggage.ContextWithBaggage(ctx, b0)

	// generic HTTP server trace
	if route == "" {
		route = r.URL.Path
	}
	opts := []trace.SpanStartOption{
		trace.WithAttributes(httpconv.ServerRequest("", r)...),
		trace.WithSpanKind(trace.SpanKindServer),
	}
	ctx, span := h.tracer.Start(ctx, route, opts...)
	defer span.End()

	attr := []metric.AddOption{metric.WithAttributes(h.meter.Attributes...)}
	e, b := api.Entry(h.decorator, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
	if b != nil {
		h.meter.TotalCount.Add(ctx, 1, attr...)
		h.meter.BlockedCount.Add(ctx, 1, attr...)
		h.meter.BlockedCountByType.Add(ctx, 1, []metric.AddOption{metric.WithAttributes(
			append(h.meter.Attributes, attribute.Key(blockedTypeKey).String(b.BlockType().String()))...)}...)

		// TODO: allow sentinel "customize block fallback" to override this 429 default
		sc := http.StatusTooManyRequests // default `429 Too Many Request`
		span.AddEvent("Stanza blocked", trace.WithAttributes(
			decoratorKey.String(h.decorator),
			blockedMessageKey.String(b.BlockMsg()),
			blockedTypeKey.String(b.BlockType().String()),
			// blockedValueKey.Int64(b.TriggeredValue()),  // how to convert interface -> int64 here?
			blockedRuleKey.String(b.TriggeredRule().String()),
		))

		logging.Error(fmt.Errorf("stanza blocked"),
			httpServerDecorator, h.decorator,
			httpServerBlockedMessage, b.BlockMsg(),
			httpServerBlockedType, b.BlockType().String(),
			httpServerBlockedValue, b.TriggeredValue(),
		)
		logging.Debug("Stanza blocked",
			httpServerBlockedRule, b.TriggeredRule().String(),
		)
		return ctx, sc
	} else {
		h.meter.TotalCount.Add(ctx, 1, attr...)
		h.meter.AllowedCount.Add(ctx, 1, attr...)

		sc := http.StatusOK
		span.AddEvent("Stanza allowed", trace.WithAttributes(
			decoratorKey.String(h.decorator),
		))
		h.propagators.Inject(ctx, propagation.HeaderCarrier(r.Header))

		e.Exit()       // cleanly exit the Sentinel Entry
		return ctx, sc // return success
	}
}
