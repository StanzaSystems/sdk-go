package stanza

import (
	"context"

	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/otel"
	"github.com/StanzaSystems/sdk-go/sentinel"
)

type Client struct {
	// Required
	// DSN or some other kind of customer key

	// Optional
	Name        string // defines applications unique name
	Release     string // defines applications version
	Environment string // defines applications environment
	StanzaHub   string // host:port (ipv4, ipv6, or resolveable hostname)
	DataSource  string // local:<path>, consul:<key>, or grpc:host:port
}

// Init initializes the SDK with ClientOptions. The returned error is
// non-nil if options is invalid, if a global client already exists, or
// if StanzaHub can't be reached.
func Init(ctx context.Context, client Client) error {
	// Set client defaults
	if client.Name == "" {
		client.Name = "unknown_service"
	}
	if client.Release == "" {
		client.Release = "0.0.0"
	}
	if client.Environment == "" {
		client.Environment = "dev"
	}
	if client.StanzaHub == "" {
		client.StanzaHub = "api.stanzahub.com"
	}
	if client.DataSource == "" {
		client.DataSource = "grpc:" + client.StanzaHub
	}

	// Initialize stanza
	global.NewState(client.Name, client.Release, client.Environment, client.StanzaHub)

	// Initialize otel
	if err := otel.Init(ctx); err != nil {
		return err
	}

	// Initialize sentinel
	if err := sentinel.Init(client.Name, client.DataSource); err != nil {
		return err
	}

	return nil
}

func NewResource(resourceName string) error {
	return global.NewResource(resourceName)
}
