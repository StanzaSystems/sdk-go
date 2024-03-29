package httphandler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/StanzaSystems/sdk-go/handlers"

	"github.com/felixge/httpsnoop"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/trace"
)

type InboundHandler struct {
	*handlers.InboundHandler
}

// NewInboundHandler returns a new InboundHandler
func NewInboundHandler(gn string, fn *string, pb *int32, dw *float32, kv *map[string]string) (*InboundHandler, error) {
	h, err := handlers.NewInboundHandler(gn, fn, pb, dw, kv)
	if err != nil {
		return nil, err
	}
	return &InboundHandler{h}, nil
}

// GuardHandlerFunc implements HTTP func(w, r) middleware for adding a Stanza Guard to an HTTP Server
func (h *InboundHandler) GuardHandlerFunction(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span, tokens := h.Start(r)
		defer span.End()

		guard := h.Guard(ctx, span, tokens)
		if guard.Blocked() {
			span.SetStatus(codes.Error, guard.BlockMessage())
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(guard.BlockMessage()))
			return
		}

		m := httpsnoop.CaptureMetrics(http.HandlerFunc(next), w, r.WithContext(ctx))
		code, msg := h.HTTPServerStatus(m.Code)
		span.SetAttributes(semconv.HTTPStatusCode(m.Code))
		span.SetStatus(code, msg)
		if code == codes.Error {
			guard.End(guard.Failure)
		} else {
			guard.End(guard.Success)
		}
	}
}

// GuardHandler implements HTTP handler (middleware) for adding a Stanza Guard to an HTTP Server
func (h *InboundHandler) GuardHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(h.GuardHandlerFunction(next.ServeHTTP))
}

func (h *InboundHandler) Start(r *http.Request) (context.Context, trace.Span, []string) {
	ctx := h.Propagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(httpconv.ServerRequest("", r)...),
	}
	if s := trace.SpanContextFromContext(ctx); s.IsValid() && s.IsRemote() {
		opts = append(opts, trace.WithLinks(trace.Link{SpanContext: s}))
	}

	ctx, span := h.Tracer().Start(ctx, r.URL.Path, opts...)
	return ctx, span, r.Header.Values("x-stanza-token")
}

// HTTPServerStatus returns a span status code and message for an HTTP status code
// value returned by a server. Status codes in the 400-499 range are not
// returned as errors.
func (h *InboundHandler) HTTPServerStatus(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 200 && code <= 299 {
		return codes.Ok, ""
	}
	if code >= 500 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}
