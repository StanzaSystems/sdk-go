package fiberstanza

import (
	"context"
	"net/http"
	"time"

	stanza "github.com/StanzaSystems/sdk-go"
	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
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
		start := time.Now()
		savedCtx, cancel := context.WithCancel(c.UserContext())

		im.ActiveRequests.Add(savedCtx, 1, im.Attributes...)
		defer func() {
			im.Duration.Record(savedCtx, float64(time.Since(start).Microseconds())/1000, im.Attributes...)
			im.RequestSize.Record(savedCtx, int64(len(c.Request().Body())), im.Attributes...)
			im.ResponseSize.Record(savedCtx, int64(len(c.Response().Body())), im.Attributes...)
			im.ActiveRequests.Add(savedCtx, -1, im.Attributes...)
			c.SetUserContext(savedCtx)
			cancel()
		}()

		// TODO(msg): implement HttpInboundHandler as fasthttp handler instead of converting to net/http?
		var req http.Request
		if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
			logging.Error(err, "failed to convert request from fasthttp")
			return c.Next() // log error and fail open
		}
		ctx, status := stanza.HttpInboundHandler(savedCtx, c.Route().Path, &im, &req)
		if status != http.StatusOK {
			return c.SendStatus(status)
		}
		c.SetUserContext(ctx)
		return c.Next()
	}
}

func Init(ctx context.Context, options ClientOptions) error {
	return stanza.Init(ctx, stanza.ClientOptions(options))
}
