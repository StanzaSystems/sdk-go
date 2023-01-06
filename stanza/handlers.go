package stanza

import (
	"context"
	"net/http"

	httphandler "github.com/StanzaSystems/sdk-go/handlers/http"
)

func InitHttpInboundMeters(decorator string) (httphandler.InboundMeters, error) {
	return httphandler.InitInboundMeters(decorator)
}

func HttpInboundHandler(ctx context.Context, decorator, route string, im *httphandler.InboundMeters, req *http.Request) (context.Context, int) {
	return httphandler.InboundHandler(ctx, gs.client.Name, decorator, route, im, req)
}
