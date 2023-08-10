package stanza

import (
	"github.com/StanzaSystems/sdk-go/handlers/httphandler"
)

// HTTP Client
func NewHttpOutboundHandler() (*httphandler.OutboundHandler, error) {
	h, err := httphandler.NewOutboundHandler(
		gs.clientOpt.APIKey,
		gs.clientId.String(),
		gs.clientOpt.Environment,
		gs.clientOpt.Name,
		OtelEnabled(),
		SentinelEnabled())
	gs.httpOutboundHandler = h
	return h, err
}

// HTTP Server
func NewHttpInboundHandler() (*httphandler.InboundHandler, error) {
	h, err := httphandler.NewInboundHandler(
		gs.clientOpt.APIKey,
		gs.clientId.String(),
		gs.clientOpt.Environment,
		gs.clientOpt.Name,
		OtelEnabled(),
		SentinelEnabled(),
	)
	gs.httpInboundHandler = h
	return h, err
}
