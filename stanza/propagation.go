package stanza

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
)

type StanzaHeaders struct{}

var (
	_ propagation.TextMapPropagator = StanzaHeaders{}

	stanzaHeaders = []string{
		"uberctx-stz-feat",
		"uberctx-stz-boost",
		"ot-baggage-stz-feat",
		"ot-baggage-stz-boost",
	}
)

// Inject sets Stanza Headers from ctx into the carrier.
func (s StanzaHeaders) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	for _, header := range stanzaHeaders {
		if ctx.Value(header) != nil {
			carrier.Set(header, ctx.Value(header).(string))
		}
	}
}

// Extract returns a copy of parent with the Stanza Headers from the carrier added.
func (s StanzaHeaders) Extract(parent context.Context, carrier propagation.TextMapCarrier) context.Context {
	for _, header := range stanzaHeaders {
		hVal := carrier.Get(header)
		if hVal != "" {
			parent = context.WithValue(parent, header, hVal)
		}
	}
	return parent
}

// Fields returns the keys who's values are set with Inject.
func (s StanzaHeaders) Fields() []string {
	return stanzaHeaders
}
