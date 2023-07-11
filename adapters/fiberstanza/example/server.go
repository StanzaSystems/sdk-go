// Example below shows how to add Stanza fault tolerance decorators
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
	"strconv"
	"strings"
	"syscall"

	"github.com/StanzaSystems/sdk-go/adapters/fiberstanza"
	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/tjarratt/babble"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	name       = "fiber-example"
	release    = "0.0.0"
	env        = "dev"
	debug      bool
	apikey     string
	listenport int
)

// For decoding ZenQuotes (https://zenquotes.io) JSON
var zq []struct {
	Q string
	A string
}

func main() {
	// Flag parsing -- most important being apikey
	flag.StringVar(&env, "environment", "dev", "Environment: for example, dev, staging, qa (default dev)")
	flag.StringVar(&apikey, "apikey", "", "(Mandatory) The API key to use with our service: obtained in the portal")
	flag.IntVar(&listenport, "listenport", 3000, "Port to listen/accept requests on")
	flag.BoolVar(&debug, "debug", true, "Debugging on/off")
	flag.Parse()

	if apikey == "" {
		fmt.Printf("Error: Mandatory API key not supplied\n")
		os.Exit(-1)
	}

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
			APIKey:      apikey,
			Name:        name,
			Release:     release,
			Environment: env,
		})
	defer stanzaExit()
	if stanzaInitErr != nil {
		logger.Fatal("stanza.init", zap.Error(stanzaInitErr))
	}

	// fiber: HTTP server
	app := fiber.New()

	// middleware: logging
	app.Use(fiberzap.New(fiberzap.Config{Logger: zap.L()}))

	// middleware: stanza inbound decorator
	app.Use(fiberstanza.New("RootDecorator"))

	// healthcheck
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Use ZenQuotes to get a random quote
	app.Get("/quote", func(c *fiber.Ctx) error {

		// Outbound request with ZenQuotes Decorator
		resp, _ :=
			fiberstanza.HttpGet(
				fiberstanza.Decorate(c, "ZenQuotes", "https://zenquotes.io/api/random"))
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

	// Get user information from GitHub
	app.Get("/github/:username", func(c *fiber.Ctx) error {
		// Set outbound request priority boost based on `X-User-Plan` request header
		opt := fiberstanza.Opt{PriorityBoost: 0, DefaultWeight: 1}
		if plan, ok := c.GetReqHeaders()["X-User-Plan"]; ok {
			if plan == "free" {
				opt.PriorityBoost -= 1
			} else if plan == "enterprise" {
				opt.PriorityBoost += 1
			}
		}

		// Optional add Headers to be sent with the outbound HTTP request
		// Here we use the GITHUB_PAT environment variable as an Authorization bearer token
		headers := make(http.Header)
		headers.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GITHUB_PAT")))
		opt.Headers = headers

		// Decorate outbound github.com request with GithubGuard
		url := fmt.Sprintf("https://api.github.com/users/%s", c.Params("username"))
		resp, err :=
			fiberstanza.HttpGet(
				fiberstanza.Decorate(c, "GithubGuard", url, opt))
		if err != nil {
			logger.Error("GithubGuard", zap.Error(err))
			if resp != nil && resp.StatusCode >= 400 {
				return c.SendStatus(resp.StatusCode)
			}
			// Use a 503 in the face of errors without an otherwise specified status code
			return c.SendStatus(http.StatusServiceUnavailable)
		}
		defer resp.Body.Close()

		// Success! 🎉
		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return c.Send(body)
		}

		// Failure. 😭
		if resp.StatusCode == http.StatusTooManyRequests {
			c.SendString("Stanza Outbound Rate Limited")
		}
		return c.SendStatus(resp.StatusCode)
	})

	// Search Google for a random word
	app.Get("/search", func(c *fiber.Ctx) error {
		// Set outbound request priority boost based on `X-User-Plan` request header
		opt := fiberstanza.Opt{PriorityBoost: 0, DefaultWeight: 1}
		if plan, ok := c.GetReqHeaders()["X-User-Plan"]; ok {
			if plan == "free" {
				opt.PriorityBoost -= 1
			} else if plan == "enterprise" {
				opt.PriorityBoost += 1
			}
		}

		// Get a random word
		babbler := babble.NewBabbler()
		babbler.Count = 1
		word := strings.TrimSuffix(babbler.Babble(), "'s")

		// Add Headers to be sent with the outbound HTTP request
		headers := make(http.Header)
		headers.Add("Referer", "https://developer.mozilla.org/en-US/docs/Web/JavaScript")
		headers.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/81.0")
		opt.Headers = headers

		// Decorate outbound google.com request with GoogleSearch
		resp, err := fiberstanza.HttpGet(
			fiberstanza.Decorate(c, "GoogleSearch",
				fmt.Sprintf("https://www.google.com/search?q=%s", word), opt))

		if err != nil {
			logger.Error("GoogleSearch", zap.Error(err))
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
			return c.SendString(fmt.Sprintf("%s %d", word, binary.Size(body)))
		}

		// Failure. 😭
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
		opt := fiberstanza.Opt{PriorityBoost: 0, DefaultWeight: 1}
		if plan, ok := c.GetReqHeaders()["X-User-Plan"]; ok {
			if plan == "free" {
				opt.PriorityBoost -= 1
			} else if plan == "enterprise" {
				opt.PriorityBoost += 1
			}
		}

		// Add Headers to be sent with the outbound HTTP request
		headers := make(http.Header)
		headers.Add("Referer", "https://gophers.slack.com/messages")
		headers.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/81.0")
		opt.Headers = headers

		// Decorate outbound request with StressTest
		resp, err :=
			fiberstanza.HttpGet(
				fiberstanza.Decorate(c, "StressTest", string(url), opt))
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
	logger.Info("Stanza example server listening ", zap.Int("port", listenport))
	listenstr := ":" + strconv.Itoa(listenport)
	go app.Listen(listenstr)

	// GRACEFUL SHUTDOWN
	// - watches for a "Done" signal to the context we setup at the start
	// - triggered by os.Interrupt, syscall.SIGINT, or syscall.SIGTERM
	<-ctx.Done()
}
