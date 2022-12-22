package fiberstanza

import (
	"context"
	"net/http"
	"time"

	"github.com/StanzaSystems/sdk-go"
	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv/v1.12.0"
)

// New creates a new fiberstanza middleware fiber.Handler
func New(config Config) fiber.Handler {
	if err := stanza.NewResource(config.ResourceName); err != nil {
		logging.Error(err, "failed to register new resource")
	}
	im, err := stanza.InitHttpInboundMeters(config.ResourceName)
	if err != nil {
		logging.Error(err, "failed to initialize new http inbound meters")
	}

	return func(c *fiber.Ctx) error {
		savedCtx, cancel := context.WithCancel(c.UserContext())
		start := time.Now()

		// TODO: do we /WANT/ any HTTP attributes on these metrics?
		// metricAttrs := httpServerMetricAttributesFromRequest(c)
		im.ActiveRequests.Add(savedCtx, 1)
		defer func() {
			im.Duration.Record(savedCtx, float64(time.Since(start).Microseconds())/1000)
			im.RequestSize.Record(savedCtx, int64(len(c.Request().Body())))
			im.ResponseSize.Record(savedCtx, int64(len(c.Response().Body())))
			im.ActiveRequests.Add(savedCtx, -1)
			c.SetUserContext(savedCtx)
			cancel()
		}()

		// TODO(msg): implement HttpInboundHandler as fasthttp handler instead of converting to net/http?
		var req http.Request
		if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
			logging.Error(err, "failed to convert request from fasthttp")
			return c.Next() // log error and fail open
		}
		if status := stanza.HttpInboundHandler(c.Context(), &im, &req); status != http.StatusOK {
			return c.SendStatus(status)
		}
		return c.Next()
	}
}

func httpServerMetricAttributesFromRequest(c *fiber.Ctx) []attribute.KeyValue {
	var attrs []attribute.KeyValue
	if c.Context().IsTLS() {
		attrs = append(attrs, semconv.HTTPSchemeHTTPS)
	} else {
		attrs = append(attrs, semconv.HTTPSchemeHTTP)
	}
	attrs = append(attrs, semconv.HTTPHostKey.String(utils.CopyString(c.Hostname())))
	attrs = append(attrs, semconv.HTTPMethodKey.String(utils.CopyString(c.Method())))
	attrs = append(attrs, semconv.HTTPRouteKey.String(utils.CopyString(c.Route().Path)))
	return attrs
}
