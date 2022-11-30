package fiberstanza

import (
	"github.com/gofiber/fiber/v2"
)

// New creates a new fiberstanza middleware handler
func New(config ...Config) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Wrap OTEL (https://github.com/gofiber/contrib/tree/main/otelfiber)

		// Wrap Sentinel (https://github.com/alibaba/sentinel-golang/blob/master/pkg/adapters/fiber/middleware.go)

		return ctx.Next()
	}
}
