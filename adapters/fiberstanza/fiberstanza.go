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
	him, err := stanza.InitHttpInboundHandler(config.ResourceName)
	if err != nil {
		logging.Error(err, "failed to initialize new http inbound handler")
	}

	return func(c *fiber.Ctx) error {
		// TODO(msg): implement HttpInboundHandler as fasthttp handler instead of converting to net/http?
		savedCtx, cancel := context.WithCancel(c.UserContext())

		start := time.Now()
		// metricAttrs := httpServerMetricAttributesFromRequest(c)
		him.ActiveRequests.Add(savedCtx, 1)
		defer func() {
			him.Duration.Record(savedCtx, float64(time.Since(start).Microseconds())/1000)
			him.RequestSize.Record(savedCtx, int64(len(c.Request().Body())))
			him.ResponseSize.Record(savedCtx, int64(len(c.Response().Body())))
			him.ActiveRequests.Add(savedCtx, -1)
			c.SetUserContext(savedCtx)
			cancel()
		}()

		var req http.Request
		if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
			logging.Error(err, "failed to convert request from fasthttp")
			return c.Next() // log error and fail open
		}
		if status := stanza.HttpInboundHandler(c.Context(), him, &req); status != http.StatusOK {
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
