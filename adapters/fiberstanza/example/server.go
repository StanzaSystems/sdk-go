// Example below shows how to add Stanza fault tolerance decorators
// to a simple Fiber service.

package main

import (
	"context"

	"github.com/StanzaSystems/sdk-go/adapters/fiberstanza"
	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	name        = "fiber-example"
	release     = "0.0.0"
	environment = "dev"
	debug       = true
)

func main() {
	ctx := context.Background()

	// Configure structured logger.
	zc := zap.NewProductionConfig()
	zc.DisableStacktrace = true
	zc.DisableCaller = true
	if environment == "dev" {
		zc.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	} else {
		zc.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	}
	if debug {
		zc.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	logger, _ := zc.Build()
	defer logger.Sync()
	zap.ReplaceGlobals(logger.WithOptions(zap.AddCallerSkip(1)))

	// Init Stanza fault tolerance library
	stanzaInitErr := fiberstanza.Init(ctx,
		fiberstanza.Client{
			Name:        name,
			Release:     release,
			Environment: environment,
			DataSource:  "local:test",
			// StanzaHub:   "host:port",
			// Logger:      zapr.NewLogger(logger.WithOptions(zap.AddCallerSkip(1))),
		})
	if stanzaInitErr != nil {
		logger.Warn("stanza.init", zap.Error(stanzaInitErr))
	}

	// fiber: HTTP server
	app := fiber.New()

	// middleware: logging
	app.Use(fiberzap.New(fiberzap.Config{Logger: zap.L()}))

	// middleware: stanza
	if stanzaInitErr == nil {
		app.Use(fiberstanza.New(fiberstanza.Decorator{
			Name: "abc",
		}))
	}

	// healthcheck
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// hello world
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	app.Listen(":3000")
}
