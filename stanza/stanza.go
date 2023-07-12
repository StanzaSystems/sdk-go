package stanza

import (
	"context"
	"errors"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type ClientOptions struct {
	// Required
	APIKey string // customer generated API key

	// Optional
	Name        string // defines applications unique name
	Release     string // defines applications version
	Environment string // defines applications environment
	StanzaHub   string // host:port (ipv4, ipv6, or resolvable hostname)
}

// Init initializes the SDK with ClientOptions. The returned error is
// non-nil if options is invalid, if a global client already exists, or
// if StanzaHub can't be reached.
func Init(ctx context.Context, co ClientOptions) (func(), error) {
	if co.APIKey == "" {
		if os.Getenv("STANZA_API_KEY") != "" {
			co.APIKey = os.Getenv("STANZA_API_KEY")
		} else {
			errMsg := "missing required Stanza API key (Hint: Set a STANZA_API_KEY environment variable!)"
			return func() {}, errors.New(errMsg)
		}
	}

	// Set client defaults
	if co.Name == "" {
		if os.Getenv("STANZA_SERVICE_NAME") != "" {
			co.Name = os.Getenv("STANZA_SERVICE_NAME")
		} else {
			co.Name = "unknown_service"
		}
	}
	if co.Release == "" {
		if os.Getenv("STANZA_SERVICE_RELEASE") != "" {
			co.Release = os.Getenv("STANZA_SERVICE_RELEASE")
		} else {
			co.Release = "0.0.0"
		}
	}
	if co.Environment == "" {
		if os.Getenv("STANZA_ENVIRONMENT") != "" {
			co.Environment = os.Getenv("STANZA_ENVIRONMENT")
		} else {
			co.Environment = "dev"
		}
	}
	if co.StanzaHub == "" {
		if os.Getenv("STANZA_HUB_ADDRESS") != "" {
			co.StanzaHub = os.Getenv("STANZA_HUB_ADDRESS")
		} else {
			co.StanzaHub = "hub.stanzasys.co:9020"
		}
	}

	// Set global propagation, we do this here since **propagation** is something
	// we want to do even if we aren't emitting OTEL metrics or traces.
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
			StanzaHeaders{}))

	// Initialize new global state
	hubDone := newState(ctx, co)

	// Return graceful shutdown function (to be deferred by the caller)
	return hubDone, nil
}

func OtelEnabled() bool {
	return os.Getenv("STANZA_NO_OTEL") == ""
}

func SentinelEnabled() bool {
	return os.Getenv("STANZA_NO_SENTINEL") == ""
}
