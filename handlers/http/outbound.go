package http

import (
	"context"
	"fmt"
	"io"
	"net/http"

	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
)

// TODO: Implement outbound/client http meters and handlers
// Use otel contrib http wrappers?
// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/net/http/otelhttp/client.go

func NewOutboundHandler(ctx context.Context, method string, url string, body io.Reader, apikey string, decoratorConfig *hubv1.DecoratorConfig, qsc hubv1.QuotaServiceClient, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	headers := ctx.Value("StanzaOutboundHeaders")
	if headers != nil {
		for k, v := range headers.(map[string]string) {
			req.Header.Set(k, v)
		}
	}
	if err != nil {
		return nil, err
	}
	if ok, token := checkQuota(apikey, decoratorConfig, qsc, tlr); ok {
		if token != "" {
			req.Header.Add("X-Stanza-Token", token)
		}
		return http.DefaultClient.Do(req)
	} else {
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
}
