package httphandler

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/keys"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/trace"
)

type OutboundHandler struct {
	*handlers.OutboundHandler
}

// NewOutboundHandler returns a new OutboundHandler
func NewOutboundHandler(gn string, fn *string, pb *int32, dw *float32, kv *map[string]string) (*OutboundHandler, error) {
	h, err := handlers.NewOutboundHandler(gn, fn, pb, dw, kv)
	if err != nil {
		return nil, err
	}
	return &OutboundHandler{h}, nil
}

// Get wraps a HTTP GET request
func (h *OutboundHandler) Get(ctx context.Context, url string) (*http.Response, error) {
	return h.Request(ctx, http.MethodGet, url, http.NoBody)
}

// Post wraps a HTTP POST request
func (h *OutboundHandler) Post(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	return h.Request(ctx, http.MethodPost, url, body)
}

// Request wraps a HTTP request of the given HTTP method
func (h *OutboundHandler) Request(ctx context.Context, httpMethod, url string, body io.Reader) (*http.Response, error) {
	if req, err := http.NewRequestWithContext(ctx, httpMethod, url, body); err != nil {
		h.FailOpen(ctx)
		return nil, err // FAIL OPEN!
	} else {
		ctx, span := h.Tracer().Start(
			req.Context(),
			h.GuardName(),
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(httpconv.ClientRequest(req)...),
		)
		defer span.End()
		h.Propagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

		guard := h.Guard(ctx, span, nil)

		// Stanza Blocked
		if guard.Blocked() {
			span.SetStatus(codes.Error, guard.BlockMessage())
			return &http.Response{
				Status:     fmt.Sprintf("%d Too Many Request", http.StatusTooManyRequests),
				StatusCode: http.StatusTooManyRequests,
				Request:    req,
				Body:       http.NoBody,
				Header:     http.Header{
					// TODO: Add retry-after header
				},
			}, guard.Error()
		}

		// Stanza Allowed
		if guard.Token() != "" {
			req.Header.Add("X-Stanza-Token", guard.Token())
		}
		if h.FeatureName() != nil {
			req.Header.Add("X-Stanza-Feature", *h.FeatureName())
		}
		if req.Header.Get("User-Agent") == "" {
			req.Header.Set("User-Agent", global.UserAgent())
		}
		if ctx.Value(keys.OutboundHeadersKey) != nil {
			for k, v := range ctx.Value(keys.OutboundHeadersKey).(http.Header) {
				req.Header.Set(k, v[0])
			}
		}
		httpClient := &http.Client{Transport: http.DefaultTransport}
		resp, err := httpClient.Do(req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			guard.End(guard.Failure)
			return resp, err // TODO: multierr with guard.Error()?
		} else {
			span.SetAttributes(
				semconv.UserAgentOriginal(req.Header.Get("User-Agent")),
				semconv.HTTPStatusCode(resp.StatusCode),
			)
			span.SetStatus(codes.Ok, "OK")
			guard.End(guard.Success)
			return resp, guard.Error()
		}
	}
}
