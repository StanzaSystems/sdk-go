// Example below shows how to add Stanza fault tolerance decorators
// to a simple net/http service.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/StanzaSystems/sdk-go/stanza"
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
	stanzaExit, stanzaInitErr := stanza.Init(ctx,
		stanza.ClientOptions{
			// APIKey:   "my-api-key", // set here or in an STANZA_API_KEY environment variable
			Name:        name,
			Release:     release,
			Environment: env,

			// optionally prefetch Guard configs
			Guard: []string{"ZenQuotes"},
		})
	defer stanzaExit()
	if stanzaInitErr != nil {
		fmt.Printf("\n%s\n\n", stanzaInitErr.Error())
		os.Exit(-1)
	}

	// healthcheck
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "OK")
	})

	// Use ZenQuotes to get a random quote
	http.HandleFunc("/quote", func(w http.ResponseWriter, r *http.Request) {

		// Create a new Stanza Guard
		stz := stanza.Guard(ctx, "ZenQuotes")

		// 🚫 Stanza Guard has *blocked* this workflow, log the reason and return 429 status
		if stz.Blocked() {
			logger.Info(stz.BlockMessage(), zap.String("reason", stz.BlockReason()))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// ✅ Stanza Guard has *allowed* this workflow, business logic goes here.
		if resp, _ := http.Get("https://zenquotes.io/api/random"); resp != nil {
			defer resp.Body.Close()

			// 🎉 Happy path, our "business logic" succeeded
			if resp.StatusCode == http.StatusOK {
				json.NewDecoder(resp.Body).Decode(&zq)
				fmt.Fprintf(w, "❝%s❞ - %s\n", zq[0].Q, zq[0].A)
				stz.End(stz.Success)
				return
			}
		}

		// 😭 Sad path, our "business logic" failed
		stz.End(stz.Failure)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	})

	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	// GRACEFUL SHUTDOWN
	// - watches for a "Done" signal to the context we setup at the start
	// - triggered by os.Interrupt, syscall.SIGINT, or syscall.SIGTERM
	<-ctx.Done()
}
