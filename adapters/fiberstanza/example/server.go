// Example below shows how to add Stanza fault tolerance decorators
// to a simple Fiber service.

package main

import (
	"context"
	"encoding/json"

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

var zq []struct {
	A string
	Q string
}

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
	shutdown, stanzaInitErr := fiberstanza.Init(ctx,
		fiberstanza.Client{
			APIKey:      "c6af1e6b-78f4-40c1-9428-2c890dcfdd7f",
			Name:        name,
			Release:     release,
			Environment: environment,
			StanzaHub:   "hub.dev.getstanza.dev:443",
			// DataSource:  "local:test",
			// Logger:      zapr.NewLogger(logger.WithOptions(zap.AddCallerSkip(1))),
		})
	defer shutdown()
	if stanzaInitErr != nil {
		logger.Error("stanza.init", zap.Error(stanzaInitErr))
	}

	// fiber: HTTP server
	app := fiber.New()

	// middleware: logging
	app.Use(fiberzap.New(fiberzap.Config{Logger: zap.L()}))

	// middleware: stanza
	if stanzaInitErr == nil {
		app.Use(fiberstanza.Middleware("RootDecorator"))
	}

	// healthcheck
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Use ZenQuotes to get a random quote
	app.Get("/", func(c *fiber.Ctx) error {
		// resp, err := http.Get("https://zenquotes.io/api/random") // before Stanza looks like this
		resp, err := fiberstanza.HttpGet("https://zenquotes.io/api/random",
			fiberstanza.Decorate("ZenQuotes", fiberstanza.GetFeatureFromContext(c)))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&zq)

		// return c.SendString("Hello, World 👋!")
		return c.SendString(zq[0].Q + " —" + zq[0].A + "\n\n")
	})

	app.Listen(":3000")
}
