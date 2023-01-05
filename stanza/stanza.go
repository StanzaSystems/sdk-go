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
	Name        string `default:"unknown_service"`        // defines applications unique name
	Release     string `default:"0.0.0"`                  // defines applications version
	Environment string `default:"dev"`                    // defines applications environment
	StanzaHub   string `default:"api.stanzahub.com"`      // host:port (ipv4, ipv6, or resolveable hostname)
	DataSource  string `default:"grpc:api.stanzahub.com"` // local:<path>, consul:<key>, or grpc:host:port
}

// Init initializes the SDK with ClientOptions. The returned error is
// non-nil if options is invalid, if a global client already exists, or
// if StanzaHub can't be reached.
func Init(ctx context.Context, client Client) error {
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
