package stanza

import (
	"context"

	httphandler "github.com/StanzaSystems/sdk-go/handlers/http"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func NewHttpOutboundHandler() (*httphandler.OutboundHandler, error) {
	h, err := httphandler.NewOutboundHandler(
		gs.clientOpt.APIKey,
		gs.clientOpt.Environment,
		gs.clientId.String(),
		OtelEnabled(),
		SentinelEnabled())
	gs.outboundHandler = h
	return h, err
}

func NewHttpInboundHandler(ctx context.Context, tlr *hubv1.GetTokenLeaseRequest) (*httphandler.InboundHandler, error) {
	md := metadata.New(map[string]string{"x-stanza-key": gs.clientOpt.APIKey})
	newCtx := metadata.NewOutgoingContext(ctx, md)
	if _, ok := gs.decoratorConfig[tlr.Selector.DecoratorName]; !ok {
		GetDecoratorConfig(newCtx, tlr.Selector.DecoratorName)
	}
	tlr.ClientId = proto.String(gs.clientId.String())
	tlr.Selector.Environment = gs.clientOpt.Environment
	ih, err := httphandler.NewInboundHandler(gs.clientOpt.APIKey, gs.decoratorConfig, tlr, OtelEnabled(), SentinelEnabled())
	gs.inboundHandlers = append(gs.inboundHandlers, ih)
	return ih, err
}
