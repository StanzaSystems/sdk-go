package stanza

import (
	"fmt"

	"github.com/StanzaSystems/sdk-go/global"
	"github.com/go-logr/logr"
)

type ClientOptions struct {
	// Required
	AppName   string `json:"appName"`
	StanzaHub string `json:"stanzaHub"` // host:port (ipv4, ipv6, or resolveable hostname)

	// Optional
	Environment string `json:"environment"`
	Logger      logr.Logger
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

	// Use supplied logger if set
	if (options.Logger != logr.Logger{}) {
		global.SetLogger(options.Logger)
	}

	// Initialize new global state
	err := global.NewState(
		options.AppName,
		options.Environment,
		options.StanzaHub)
	if err != nil {
		return err
	}
	return nil
}

// SetLogger configures the logger used internally by the SDK. This allows you
// to "Bring Your Own Logger" (by way of the go-logr/logr logging API).
//
// It can also be passed in as an option to Init().
func SetLogger(logger logr.Logger) {
	global.SetLogger(logger)
}
