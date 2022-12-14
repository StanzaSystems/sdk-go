package stanza

import (
	"context"
	"fmt"

	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/otel"
	"github.com/StanzaSystems/sdk-go/sentinel"
)

type ClientOptions struct {
	Name        string // defines applications name --REQUIRED--
	Release     string // defines applications version (default: v0.0.0)
	Environment string // defines applications environment (default: dev)

	// TODO: figure out if we need this?
	StanzaHub string // host:port (ipv4, ipv6, or resolveable hostname)

	// TODO: make sentinel.DataSourceOptions an interface?
	DataSource sentinel.DataSourceOptions // sentinel datasource to get flowcontrol rules from
}

// Init initializes the SDK with ClientOptions. The returned error is
// non-nil if options is invalid, if a global client already exists, or
// if StanzaHub can't be reached.
func Init(ctx context.Context, options ClientOptions) error {
	// Check for required options
	if options.StanzaHub == "" {
		return fmt.Errorf("StanzaHub is a required option")
	}

	// Initialize stanza global state
	global.NewState(options.Name, options.Release, options.Environment, options.StanzaHub)

	// Initialize otel
	if err := otel.Init(ctx); err != nil {
		return err
	}

	// Initialize sentinel
	if err := sentinel.Init(options.Name, options.DataSource); err != nil {
		return err
	}

	return nil
}

func NewResource(resourceName string) error {
	return global.NewResource(resourceName)
}
