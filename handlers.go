package stanza

import (
	"context"
	"net/http"

	httphandler "github.com/StanzaSystems/sdk-go/handlers/http"
)

func InitHttpInboundMeters(res string) (httphandler.InboundMeters, error) {
	return httphandler.InitInboundMeters(res)
}

func HttpInboundHandler(ctx context.Context, im *httphandler.InboundMeters, req *http.Request) int {
	return httphandler.InboundHandler(ctx, im, req)
}
