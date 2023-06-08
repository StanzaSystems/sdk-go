package stanza

import (
	"context"
	"io"
	"net/http"

	httphandler "github.com/StanzaSystems/sdk-go/handlers/http"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"google.golang.org/protobuf/proto"
)

func NewHttpOutboundHandler(ctx context.Context, method string, url string, body io.Reader, tlr *hubv1.GetTokenLeaseRequest) (*http.Response, error) {
	if _, ok := gs.decoratorConfig[tlr.Selector.DecoratorName]; !ok {
		GetDecoratorConfig(ctx, tlr.Selector.DecoratorName)
	}
	tlr.ClientId = proto.String(gs.clientId.String())
	tlr.Selector.Environment = gs.clientOpt.Environment
	return httphandler.NewOutboundHandler(ctx, method, url, body, gs.decoratorConfig[tlr.Selector.DecoratorName], tlr)
}

func NewHttpInboundHandler(ctx context.Context, decorator string) (*httphandler.InboundHandler, error) {
	if _, ok := gs.decoratorConfig[decorator]; !ok {
		GetDecoratorConfig(ctx, decorator)
	}
	return httphandler.NewInboundHandler(decorator, gs.decoratorConfig, OtelEnabled(), SentinelEnabled())
}
