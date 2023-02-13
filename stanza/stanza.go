package stanza

import (
	"context"
	"errors"

	"github.com/StanzaSystems/sdk-go/otel"
	"github.com/StanzaSystems/sdk-go/sentinel"
)

type ClientOptions struct {
	// Required
	APIKey string // customer generated API key

	// Optional
	Name        string // defines applications unique name
	Release     string // defines applications version
	Environment string // defines applications environment
	StanzaHub   string // host:port (ipv4, ipv6, or resolvable hostname)
	DataSource  string // local:<path>, consul:<key>, or grpc:host:port
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
		co.StanzaHub = "api.stanzahub.com"
	}

	// Initialize stanza
	newState(co)

	// TODO: register call to stanza hub API should return otel-collector address

	// Initialize otel
	shutdown, err := otel.Init(ctx, gs.client.Name, gs.client.Release, gs.client.Environment)
	if err != nil {
		return func() { shutdown() }, err
	}

	// Initialize sentinel
	if co.DataSource != "" {
		if err := sentinel.Init(gs.client.Name, gs.client.DataSource); err != nil {
			return func() { shutdown() }, err
		}
	}

	// Return OTEL shutdown (to be deferred by the caller)
	return func() { shutdown() }, nil
}

func NewDecorator(name string) error {
	// TODO(msg): register new decorator with stanzahub
	return nil
}
