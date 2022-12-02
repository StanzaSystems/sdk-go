package fiberstanza

import (
	"github.com/StanzaSystems/sdk-go/global"
	"github.com/gofiber/fiber/v2"
)

// New creates a new fiberstanza middleware fiber.Handler
func New(config Config) fiber.Handler {
	if config.ResourceName != "" {
		err := global.NewResource(config.ResourceName)
		if err != nil {
			global.Error(err, "failed to register new resource")
		}
	}

	return func(ctx *fiber.Ctx) error {
		global.Debug("fiberstanza function start")

		// Wrap OTEL (https://github.com/gofiber/contrib/tree/main/otelfiber)

		// Wrap Sentinel (https://github.com/alibaba/sentinel-golang/blob/master/pkg/adapters/fiber/middleware.go)

		global.Debug("fiberstanza function finish")
		return ctx.Next()
	}
}
