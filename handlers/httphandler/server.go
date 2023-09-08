package httphandler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/keys"

	"github.com/felixge/httpsnoop"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/trace"
)

type InboundHandler struct {
	*handlers.InboundHandler
	guardName     string
	featureName   *string  // overrides request baggage
	priorityBoost *int32   // overrides request baggage
	defaultWeight *float32 // overrides request baggage
}

// NewInboundHandler returns a new InboundHandler
func NewInboundHandler(gn string, fn *string, pb *int32, dw *float32) (*InboundHandler, error) {
	h, err := handlers.NewInboundHandler()
	if err != nil {
		return nil, err
	}
	return &InboundHandler{h, gn, fn, pb, dw}, nil
}

// Guard implements HTTP handler (middleware) for adding a Stanza Guard to an HTTP Server
func (h *InboundHandler) Guard(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx, span, tokens := h.Start(r)
		defer span.End()

		guard := h.NewGuard(ctx, span, h.guardName, tokens)
		if guard.Blocked() {
			span.SetStatus(codes.Error, guard.BlockMessage())
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(guard.BlockMessage()))
			return
		}

		m := httpsnoop.CaptureMetrics(next, w, r.WithContext(ctx))
		span.SetAttributes(semconv.HTTPStatusCode(m.Code))
		span.SetStatus(h.HTTPServerStatus(m.Code))
		if m.Code != http.StatusOK {
			guard.End(guard.Failure)
		} else {
			guard.End(guard.Success)
		}
	}
	return http.HandlerFunc(fn)
}

func (h *InboundHandler) Start(r *http.Request) (context.Context, trace.Span, []string) {
	ctx := h.Propagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	if h.featureName != nil {
		ctx = context.WithValue(ctx, keys.StanzaFeatureNameKey, *h.featureName)
	}
	if h.priorityBoost != nil {
		ctx = context.WithValue(ctx, keys.StanzaPriorityBoostKey, *h.priorityBoost)
	}
	if h.defaultWeight != nil {
		ctx = context.WithValue(ctx, keys.StanzaDefaultWeightKey, *h.defaultWeight)
	}

	ctx, span := h.Tracer().Start(
		ctx,
		r.URL.Path,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(httpconv.ServerRequest("", r)...),
	)
	return ctx, span, r.Header.Values("x-stanza-token")
}

// HTTPServerStatus returns a span status code and message for an HTTP status code
// value returned by a server. Status codes in the 400-499 range are not
// returned as errors.
func (h *InboundHandler) HTTPServerStatus(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 500 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}
