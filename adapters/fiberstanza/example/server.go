// Example below shows how to add Stanza fault tolerance guards
// to a simple Fiber service.

package main

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

var (
	name    = "fiber-example"
	release = "1.0.0"
	env     string
	debug   bool
	port    int
)

// For decoding ZenQuotes (https://zenquotes.io) JSON
var zq []struct {
	Q string
	A string
}

func main() {
	// Flag parsing -- most important being apikey
	flag.StringVar(&env, "environment", "dev", "Environment: for example, dev, staging, qa (default dev)")
	flag.IntVar(&port, "port", 3000, "Port to listen/accept requests on")
	flag.BoolVar(&debug, "debug", true, "Debugging on/off")
	flag.Parse()

	// Create an interruptible context to use for graceful server shutdowns
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Configure structured logger.
	zc := zap.NewProductionConfig()
	zc.DisableStacktrace = true
	zc.DisableCaller = true
	if env == "dev" {
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
			// APIKey:   "my-api-key", // set here or in an STANZA_API_KEY environment variable
			Name:        name,
			Release:     release,
			Environment: env,

			// optionally prefetch Guard configs
			Guard: []string{"StressTest"},
		})
	defer stanzaExit()
	if stanzaInitErr != nil {
		fmt.Printf("\n%s\n\n", stanzaInitErr.Error())
		os.Exit(-1)
	}

	// fiber: HTTP server
	app := fiber.New()

	// service release version
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString(release)
	})

	// healthcheck
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// middleware: logging
	app.Use(fiberzap.New(fiberzap.Config{Logger: zap.L()}))

	// middleware: stanza inbound guard
	// app.Use(fiberstanza.New("RootGuard"))

	// Use ZenQuotes to get a random quote
	app.Get("/quote", func(c *fiber.Ctx) error {

		// Outbound request with ZenQuotes Guard
		resp, err := fiberstanza.HttpGet(c, "ZenQuotes", "https://zenquotes.io/api/random")
		if err != nil {
			logger.Error("ZenQuotes", zap.Error(err))
		}
		defer resp.Body.Close()

		// Success! 🎉
		// Our outbound HTTP request succeeded, this is the "happy path"!
		if resp.StatusCode == http.StatusOK {
			json.NewDecoder(resp.Body).Decode(&zq)
			return c.SendString("❝" + zq[0].Q + "❞ -" + zq[0].A + "\n")
		}

		// Failure. 😭
		// Consider how you want to handle this case! This could be a "429 Too Many Requests"
		// (which we check for explicitly) or it could be a transient 5xx. Either way we don't
		// have the response we were hoping for and we have to decide how to handle it. You might
		// consider displaying a user friendly "Something went wrong!" message, or if this is an
		// optional component of a larger page, just skip rendering it.
		//
		// For example purposes we send a custom message in the body if it was rate limited.
		if resp.StatusCode == http.StatusTooManyRequests {
			c.SendString("Stanza Outbound Rate Limited")
		}
		return c.SendStatus(resp.StatusCode)
	})

	// Test a given URL to see when it rate limits
	app.Get("/test/:url", func(c *fiber.Ctx) error {
		url, err := base64.StdEncoding.DecodeString(c.Params("url"))
		if err != nil {
			logger.Error("StressTest", zap.Error(err))
			return c.SendStatus(http.StatusTeapot)
		}

		// Set outbound request priority boost based on `X-User-Plan` request header
		opt := fiberstanza.Opt{PriorityBoost: 0}
		if plan, ok := c.GetReqHeaders()["X-User-Plan"]; ok {
			if plan == "free" {
				opt.PriorityBoost -= 1
			} else if plan == "enterprise" {
				opt.PriorityBoost += 1
			}
		}

		// Guard outbound request with StressTest
		resp, err := fiberstanza.HttpGet(c, "StressTest", string(url), opt)
		if err != nil {
			logger.Error("StressTest", zap.Error(err))
			if resp != nil && resp.StatusCode >= 400 {
				return c.SendStatus(resp.StatusCode)
			}
			// Use a 503 in the face of errors without a proper error code
			return c.SendStatus(http.StatusServiceUnavailable)
		}
		defer resp.Body.Close()

		// Success! 🎉
		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return c.SendString(fmt.Sprint(binary.Size(body)))
		}

		// Failure. 😭
		if resp.StatusCode == http.StatusTooManyRequests {
			c.SendString("Stanza Outbound Rate Limited")
		}
		return c.SendStatus(resp.StatusCode)
	})
	go app.Listen(fmt.Sprintf(":%d", port))

	// GRACEFUL SHUTDOWN
	// - watches for a "Done" signal to the context we setup at the start
	// - triggered by os.Interrupt, syscall.SIGINT, or syscall.SIGTERM
	<-ctx.Done()
}
