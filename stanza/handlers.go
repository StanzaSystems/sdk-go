package stanza

import (
	"context"
	"net/http"

	httphandler "github.com/StanzaSystems/sdk-go/handlers/http"
)

func InitHttpInboundMeters(res string) (httphandler.InboundMeters, error) {
	return httphandler.InitInboundMeters(res)
}

func HttpInboundHandler(ctx context.Context, route string, im *httphandler.InboundMeters, req *http.Request) (context.Context, int) {
	return httphandler.InboundHandler(ctx, route, im, req)
}
