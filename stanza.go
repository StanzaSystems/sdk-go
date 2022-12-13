package stanza

import (
	"fmt"

	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/sentinel"
)

type ClientOptions struct {
	// Required
	AppName   string // name of this application
	StanzaHub string // host:port (ipv4, ipv6, or resolveable hostname)

	// TODO: make sentinel.DataSourceOptions an interface?
	DataSource sentinel.DataSourceOptions // sentinel datasource to get flowcontrol rules from

	// Optional
	Environment string      // any string which defines your environment (default: dev)
}

// Init initializes the SDK with ClientOptions. The returned error is
// non-nil if options is invalid, if a global client already exists, or
// if StanzaHub can't be reached.
func Init(options ClientOptions) error {
	// Check for required options
	if options.AppName == "" {
		return fmt.Errorf("AppName is a required option")
	}
	if options.StanzaHub == "" {
		return fmt.Errorf("StanzaHub is a required option")
	}

	// Apply option defaults if unset
	if options.Environment == "" {
		options.Environment = "dev"
	}

	// Initialize stanza global state
	if err := global.NewState(options.AppName, options.Environment, options.StanzaHub); err != nil {
		return err
	}

	// Initialize otel?

	// Initialize sentinel
	if err := sentinel.Init(options.AppName, options.DataSource); err != nil {
		return err
	}

	return nil
}

func NewResource(resourceName string) error {
	return global.NewResource(resourceName)
}
