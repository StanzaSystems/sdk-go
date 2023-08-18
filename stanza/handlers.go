package stanza

import (
	"context"
	"fmt"

	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/handlers/httphandler"
	"github.com/StanzaSystems/sdk-go/logging"
	"go.opentelemetry.io/otel/trace"
)

var (
	gh *handlers.Handler = nil
)

// HTTP Client
func NewHttpOutboundHandler() (*httphandler.OutboundHandler, error) {
	return httphandler.NewOutboundHandler()
}

// HTTP Server
func NewHttpInboundHandler() (*httphandler.InboundHandler, error) {
	return httphandler.NewInboundHandler()
}

// Generic Guard
func Guard(ctx context.Context, name string) *handlers.Guard {
	if gh == nil {
		var err error
		gh, err = handlers.NewHandler()
		if err != nil {
			err = fmt.Errorf("failed to create guard handler: %s", err)
			logging.Error(err)
			return gh.NewGuardError(ctx, nil, nil, err)
		}
	}
	opts := []trace.SpanStartOption{
		// WithAttributes?
		trace.WithSpanKind(trace.SpanKindInternal),
	}
	ctx, span := gh.Tracer().Start(ctx, name, opts...)
	defer span.End()
	return gh.NewGuard(ctx, span, name, []string{})
}
