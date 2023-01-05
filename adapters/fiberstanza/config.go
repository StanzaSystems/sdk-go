package fiberstanza

import "github.com/StanzaSystems/sdk-go/sentinel"

// Config defines the config for fiberstanza middleware.
type Config struct {
	ResourceName string `json:"resourceName"` // optional (but required if you want to protect multiple resources)
}

// ClientOptions defines the config for Stanza
type ClientOptions struct {
	Name        string // defines applications name --REQUIRED--
	Release     string // defines applications version (default: v0.0.0)
	Environment string // defines applications environment (default: dev)

	// TODO: figure out if we need this?
	StanzaHub string // host:port (ipv4, ipv6, or resolveable hostname)

	// TODO: make sentinel.DataSourceOptions an interface?
	DataSource sentinel.DataSourceOptions // sentinel datasource to get flowcontrol rules from
}
