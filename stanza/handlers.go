package stanza

import (
	httphandler "github.com/StanzaSystems/sdk-go/handlers/http"
)

func NewHttpOutboundHandler() (*httphandler.OutboundHandler, error) {
	h, err := httphandler.NewOutboundHandler(
		gs.clientOpt.APIKey,
		gs.clientId.String(),
		gs.clientOpt.Environment,
		gs.clientOpt.Name,
		OtelEnabled(),
		SentinelEnabled())
	gs.outboundHandler = h
	return h, err
}

func NewHttpInboundHandler() (*httphandler.InboundHandler, error) {
	h, err := httphandler.NewInboundHandler(
		gs.clientOpt.APIKey,
		gs.clientId.String(),
		gs.clientOpt.Environment,
		gs.clientOpt.Name,
		OtelEnabled(),
		SentinelEnabled())
	gs.inboundHandler = h
	return h, err
}
