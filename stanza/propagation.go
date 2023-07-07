package stanza

import (
	"context"

	"github.com/StanzaSystems/sdk-go/keys"

	"go.opentelemetry.io/otel/propagation"
)

type StanzaHeaders struct{}

var (
	_ propagation.TextMapPropagator = StanzaHeaders{}

	stanzaHeaders = []keys.ContextKey{
		keys.UberctxStzBoostKey,
		keys.UberctxStzFeatKey,
		keys.OtStzBoostKey,
		keys.OtStzFeatKey,
	}
)

// Inject sets Stanza Headers from ctx into the carrier.
func (s StanzaHeaders) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	for _, key := range stanzaHeaders {
		if ctx.Value(key) != nil {
			carrier.Set(string(key), ctx.Value(key).(string))
		}
	}
}

// Extract returns a copy of parent with the Stanza Headers from the carrier added.
func (s StanzaHeaders) Extract(parent context.Context, carrier propagation.TextMapCarrier) context.Context {
	for _, key := range stanzaHeaders {
		hVal := carrier.Get(string(key))
		if hVal != "" {
			parent = context.WithValue(parent, key, hVal)
		}
	}
	return parent
}

// Fields returns the keys who's values are set with Inject.
func (s StanzaHeaders) Fields() (headers []string) {
	for _, key := range stanzaHeaders {
		headers = append(headers, string(key))
	}
	return headers
}
