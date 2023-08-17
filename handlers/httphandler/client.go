package httphandler

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/keys"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type OutboundHandler struct {
	*handlers.OutboundHandler
}

// NewOutboundHandler returns a new OutboundHandler
func NewOutboundHandler() (*OutboundHandler, error) {
	h, err := handlers.NewOutboundHandler()
	if err != nil {
		return nil, err
	}
	return &OutboundHandler{h}, nil
}

// Get wraps a HTTP GET request
func (h *OutboundHandler) Get(ctx context.Context, guard, url string) (*http.Response, error) {
	return h.Request(ctx, guard, http.MethodGet, url, http.NoBody)
}

// Post wraps a HTTP POST request
func (h *OutboundHandler) Post(ctx context.Context, guard, url string, body io.Reader) (*http.Response, error) {
	return h.Request(ctx, guard, http.MethodPost, url, body)
}

// Request wraps a HTTP request of the given HTTP method
func (h *OutboundHandler) Request(ctx context.Context, guardName, httpMethod, url string, body io.Reader) (*http.Response, error) {
	// TODO: Add a Span around this Request, like otelhttp does:
	// https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp#Get
	if req, err := http.NewRequestWithContext(ctx, httpMethod, url, body); err != nil {
		return nil, err // FAIL OPEN!
	} else {
		guard := h.NewGuard(ctx, guardName)

		// Stanza Blocked
		if guard.Blocked() {
			return &http.Response{
				Status:     fmt.Sprintf("%d Too Many Request", http.StatusTooManyRequests),
				StatusCode: http.StatusTooManyRequests,
				Request:    req,
				Body:       http.NoBody,
				Header:     http.Header{
					// TODO: Add retry-after header
				},
			}, nil
		}

		// Stanza Allowed
		if guard.Token() != "" {
			req.Header.Add("X-Stanza-Token", guard.Token())
		}
		if req.Header.Get("User-Agent") == "" {
			req.Header.Set("User-Agent", global.UserAgent())
		}
		if ctx.Value(keys.OutboundHeadersKey) != nil {
			for k, v := range ctx.Value(keys.OutboundHeadersKey).(http.Header) {
				req.Header.Set(k, v[0])
			}
		}
		httpClient := &http.Client{
			Transport: otelhttp.NewTransport(
				http.DefaultTransport,
			)}
		resp, err := httpClient.Do(req)
		if err != nil {
			guard.End(guard.Failure)
		} else {
			guard.End(guard.Success)
		}
		return resp, err
	}
}
