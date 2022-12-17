package fiberstanza

import (
	"context"
	"net/http"
	"time"

	"github.com/StanzaSystems/sdk-go"
	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/propagation"
)

const (
	instrumentationName = "github.com/StanzaSystems/sdk-go/adapters/fiberstanza"

	metricNameHttpServerDuration       = "http.server.duration"
	metricNameHttpServerRequestSize    = "http.server.request.size"
	metricNameHttpServerResponseSize   = "http.server.response.size"
	metricNameHttpServerActiveRequests = "http.server.active_requests"
)

// New creates a new fiberstanza middleware fiber.Handler
func New(config Config) fiber.Handler {
	if err := stanza.NewResource(config.ResourceName); err != nil {
		logging.Error(err, "failed to register new resource")
	}

	meter := global.Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(contrib.SemVersion()),
	)
	httpServerDuration, err := meter.SyncFloat64().Histogram(
		metricNameHttpServerDuration,
		instrument.WithUnit(unit.Milliseconds),
		instrument.WithDescription("measures the duration inbound HTTP requests"))
	if err != nil {
		otel.Handle(err)
	}
	httpServerRequestSize, err := meter.SyncInt64().Histogram(
		metricNameHttpServerRequestSize,
		instrument.WithUnit(unit.Bytes),
		instrument.WithDescription("measures the size of HTTP request messages"))
	if err != nil {
		otel.Handle(err)
	}
	httpServerResponseSize, err := meter.SyncInt64().Histogram(
		metricNameHttpServerResponseSize,
		instrument.WithUnit(unit.Bytes),
		instrument.WithDescription("measures the size of HTTP response messages"))
	if err != nil {
		otel.Handle(err)
	}
	httpServerActiveRequests, err := meter.SyncInt64().UpDownCounter(
		metricNameHttpServerActiveRequests,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of concurrent HTTP requests that are currently in-flight"))
	if err != nil {
		otel.Handle(err)
	}

	return func(c *fiber.Ctx) error {
		savedCtx, cancel := context.WithCancel(c.UserContext())

		start := time.Now()
		httpServerActiveRequests.Add(savedCtx, 1)
		defer func() {
			httpServerDuration.Record(savedCtx, float64(time.Since(start).Microseconds())/1000)
			httpServerRequestSize.Record(savedCtx, int64(len(c.Request().Body())))
			httpServerResponseSize.Record(savedCtx, int64(len(c.Response().Body())))
			httpServerActiveRequests.Add(savedCtx, -1)
			c.SetUserContext(savedCtx)
			cancel()
		}()

		reqHeader := make(http.Header)
		c.Request().Header.VisitAll(func(k, v []byte) {
			reqHeader.Add(string(k), string(v))
		})

		prop := otel.GetTextMapPropagator()
		ctx := prop.Extract(savedCtx, propagation.HeaderCarrier(reqHeader))

		// TODO: Start the trace here
		// spanName := utils.CopyString(c.Path())
		// ctx, span := tracer.Start(ctx, spanName, opts...)
		// defer span.End()

		// pass the span through userContext
		c.SetUserContext(ctx)

		// TODO(msg): implement HttpInboundHandler as fasthttp handler instead of converting to net/http?
		var req http.Request
		if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
			return err
		}
		if status := stanza.HttpInboundHandler(config.ResourceName, &req); status != http.StatusOK {
			return c.SendStatus(status)
		}
		return c.Next()
	}
}
