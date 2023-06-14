// Example below shows how to add Stanza fault tolerance decorators
// to a simple Fiber service.

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

// For decoding ZenQuotes (https://zenquotes.io) JSON
var zq []struct {
	Q string
	A string
}

func main() {
	// Create an interruptible context to use for graceful server shutdowns
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
	stanzaExit, stanzaInitErr := fiberstanza.Init(ctx,
		fiberstanza.Client{
			APIKey:      "c6af1e6b-78f4-40c1-9428-2c890dcfdd7f",
			Name:        name,
			Release:     release,
			Environment: environment,
		})
	defer stanzaExit()
	if stanzaInitErr != nil {
		logger.Fatal("stanza.init", zap.Error(stanzaInitErr))
	}

	// fiber: HTTP server
	app := fiber.New()

	// middleware: logging
	app.Use(fiberzap.New(fiberzap.Config{Logger: zap.L()}))

	// middleware: stanza
	app.Use(fiberstanza.Middleware(ctx, "RootDecorator"))

	// healthcheck
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Use ZenQuotes to get a random quote
	app.Get("/", func(c *fiber.Ctx) error {
		// resp, err := http.Get("https://zenquotes.io/api/random") // before Stanza looks like this
		resp, _ := fiberstanza.HttpGet(ctx, "https://zenquotes.io/api/random",
			fiberstanza.Decorate("ZenQuotes", fiberstanza.GetFeatureFromContext(c)))

		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			// Success! 🎉
			// Our outbound HTTP request succeeded, this is the "happy path"!
			json.NewDecoder(resp.Body).Decode(&zq)
			return c.SendString("❝" + zq[0].Q + "❞ -" + zq[0].A + "\n")
		}

		// Failure. 😭
		// Consider how you want to handle this case! This could be a "429 Too Many Requests"
		// (you can check for this explicitly) or it could be a transient 5xx. Either way we don't
		// have the response we were hoping for and we have to decide how to handle it. You might
		// consider displaying a user friendly "Something went wrong!" message, or if this is an
		// optional component of a larger page, just skip rendering it.
		//
		// For example purposes we simple return the status code here.
		c.SendString("Stanza Outbound Rate Limited")
		return c.SendStatus(resp.StatusCode)
	})

	go app.Listen(":3000")

	// GRACEFUL SHUTDOWN
	// - watches for a "Done" signal to the context we setup at the start
	// - triggered by os.Interrupt, syscall.SIGINT, or syscall.SIGTERM
	<-ctx.Done()
}
