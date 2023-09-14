package otel

import (
	"context"
	"net/http"

	"github.com/StanzaSystems/sdk-go/keys"
	"google.golang.org/grpc/metadata"

	"go.opentelemetry.io/otel/propagation"
)

type StanzaHeaders struct{}

var (
	// Assert that StanzaHeaders implements the TextMapPropagator interface
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

// Assert that metadataSupplier implements the TextMapCarrier interface
var _ propagation.TextMapCarrier = &MetadataSupplier{}

type MetadataSupplier struct {
	Metadata *metadata.MD
}

// Set returns value for given key from metadata.
func (s *MetadataSupplier) Get(key string) string {
	values := s.Metadata.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set sets key-value pairs in metadata.
func (s *MetadataSupplier) Set(key string, value string) {
	s.Metadata.Set(key, value)
}

// Keys returns the keys who's values are set with Inject.
func (s *MetadataSupplier) Keys() []string {
	out := make([]string, 0, len(*s.Metadata))
	for key := range *s.Metadata {
		out = append(out, key)
	}
	return out
}

// Helper function which extracts and propagates OTEL TraceContext, Baggage, and StanzaHeaders
// from a given http.Request.
func ContextWithHeaders(r *http.Request) context.Context {
	return GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
}
