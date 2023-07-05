package stanza

import (
	"context"
	"errors"
	"os"
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
		return func() {}, errors.New("missing required Stanza API key")
	}

	// Set client defaults
	if co.Name == "" {
		co.Name = "unknown_service"
	}
	if co.Release == "" {
		co.Release = "0.0.0"
	}
	if co.Environment == "" {
		co.Environment = "dev"
	}
	if co.StanzaHub == "" {
		co.StanzaHub = "hub.getstanza.io:9020"
	}

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
