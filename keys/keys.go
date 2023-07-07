package keys

type ContextKey string

const (
	StzBoost = "stz-boost"
	StzFeat  = "stz-feat"
)

var (
	OutboundHeadersKey = ContextKey("stanza-outbound-headers")
	UberctxStzBoostKey = ContextKey("uberctx-" + StzBoost)
	UberctxStzFeatKey  = ContextKey("uberctx-" + StzFeat)
	OtStzBoostKey      = ContextKey("ot-baggage-" + StzBoost)
	OtStzFeatKey       = ContextKey("ot-baggage-" + StzFeat)
)
