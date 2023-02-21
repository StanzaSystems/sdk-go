package http

import (
	"net/http"
)

// TODO: Implement outbound/client http meters and handlers

// Use otel contrib http wrappers?
// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/net/http/otelhttp/client.go

func HTTPGetHandler(url, decorator, feature string) (*http.Response, error) {
	// TODO: Get DecoratorConfig (if we don't already have it, also setup a background poller)

	// Make /v1/quota/token request -- if err != nil
	// return response of token request, else
	return http.Get(url)
}
