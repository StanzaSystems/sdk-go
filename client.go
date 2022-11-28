package stanza

type ClientOptions struct {
	StanzaHub string // host:port (ipv4, ipv6, or resolveable hostname)
}

// Client is the underlying processor that is used by the main API.
// It must be created with NewClient.
type Client struct {
	options ClientOptions
}

// NewClient creates and returns an instance of Client configured using
// ClientOptions.
//
// Most users will not create clients directly. Instead, initialize the
// SDK with stanza.Init().
func NewClient(options ClientOptions) (*Client, error) {
	client := Client{
		options: options,
	}

	return &client, nil
}
