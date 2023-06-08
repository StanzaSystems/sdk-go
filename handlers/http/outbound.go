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

func NewOutboundHandler(ctx context.Context, method string, url string, body io.Reader, decoratorConfig *hubv1.DecoratorConfig, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	if decoratorConfig.GetCheckQuota() {
		fmt.Println("TODO: Enable OutboundHandler QuotaChecks")
		// hubv1.QuotaServiceClient.GetTokenLease(ctx, tlr)
		// Make /v1/quota/tokenlease request with tlr  -- if err != nil
		// return response of token request, else
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}
