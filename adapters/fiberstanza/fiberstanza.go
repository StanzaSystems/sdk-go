package fiberstanza

import (
	"net/http"

	"github.com/StanzaSystems/sdk-go"
	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// New creates a new fiberstanza middleware fiber.Handler
func New(config Config) fiber.Handler {
	if err := stanza.NewResource(config.ResourceName); err != nil {
		logging.Error(err, "failed to register new resource")
	}

	return func(ctx *fiber.Ctx) error {
		// TODO(msg): implement HttpInboundHandler as fasthttp handler instead of converting to net/http
		var req http.Request
		if err := fasthttpadaptor.ConvertRequest(ctx.Context(), &req, true); err != nil {
			return err
		}
		if err := stanza.HttpInboundHandler(config.ResourceName, &req); err != nil {
			return err
		}
		return ctx.Next()
	}
}
