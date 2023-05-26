package stanza

import (
	"context"
	"errors"
	"log"

	"github.com/StanzaSystems/sdk-go/otel"
	"github.com/StanzaSystems/sdk-go/sentinel"
	"google.golang.org/grpc"
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
		co.StanzaHub = "hub.getstanza.io"
	}

	// Initialize stanza
	newState(co)

	// TODO: register call to stanza hub API should return otel-collector address

	// TODO: register call to stanza hub to swap API key for otel bearer token
	token, err := GetBearerToken(co.StanzaHub, co.APIKey)
	if err != nil {
		return func() {}, err
	}

	// Initialize otel
	shutdown, err := otel.Init(ctx, gs.client.Name, gs.client.Release, gs.client.Environment, token)
	if err != nil {
		return func() { shutdown() }, err
	}

	// Initialize sentinel
	if SentinelEnabled() {
		if err := sentinel.Init(gs.client.Name, gs.client.DataSource); err != nil {
			return func() { shutdown() }, err
		}
	}

	// Return OTEL shutdown (to be deferred by the caller)
	return func() { shutdown() }, nil
}
