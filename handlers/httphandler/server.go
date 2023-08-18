package httphandler

import (
	"context"
	"net/http"

	"github.com/StanzaSystems/sdk-go/handlers"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/trace"
)

type InboundHandler struct {
	*handlers.InboundHandler
}

// NewInboundHandler returns a new InboundHandler
func NewInboundHandler() (*InboundHandler, error) {
	h, err := handlers.NewInboundHandler()
	if err != nil {
		return nil, err
	}
	return &InboundHandler{h}, nil
}

func (h *InboundHandler) VerifyServingCapacity(r *http.Request, route string, guardName string) (context.Context, int) {
	ctx := h.Propagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	opts := []trace.SpanStartOption{
		trace.WithAttributes(httpconv.ServerRequest("", r)...),
		trace.WithSpanKind(trace.SpanKindServer),
	}
	ctx, span := h.Tracer().Start(ctx, guardName, opts...)
	defer span.End()

	guard := h.NewGuard(ctx, span, guardName, r.Header.Values("x-stanza-token"))

	// Stanza Blocked
	if guard.Blocked() {
		return ctx, http.StatusTooManyRequests
	}

	// Stanza Allowed
	// TODO: We need to execute "Next" and check it's response code
	// in order to decide Success vs Failure
	guard.End(guard.Success) // TODO: FIX
	return ctx, http.StatusOK
}
