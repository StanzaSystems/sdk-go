package keys

type outboundHeadersContextKey string

var (
	OutboundHeadersKey = outboundHeadersContextKey("stanzaOutboundHeaders")
)
