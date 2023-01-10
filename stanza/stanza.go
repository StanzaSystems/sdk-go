package stanza

import (
	"context"

	"github.com/StanzaSystems/sdk-go/otel"
	"github.com/StanzaSystems/sdk-go/sentinel"
)

type ClientOptions struct {
	// Required
	// DSN or some other kind of customer key

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
func Init(ctx context.Context, co ClientOptions) error {
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
	if co.DataSource == "" {
		co.DataSource = "grpc:" + co.StanzaHub
	}

	// Initialize stanza
	newState(co)

	// Initialize otel
	if err := otel.Init(ctx, gs.client.Name, gs.client.Release, gs.client.Environment); err != nil {
		return err
	}

	// Initialize sentinel
	if err := sentinel.Init(gs.client.Name, gs.client.DataSource); err != nil {
		return err
	}

	return nil
}

func NewDecorator(name string) error {
	// TODO(msg): register new decorator with stanzahub
	return nil
}
