package stanza

import (
	"github.com/StanzaSystems/sdk-go/handlers/http/httpclient"
	"github.com/StanzaSystems/sdk-go/handlers/http/httpserver"
)

// HTTP Client
func NewHttpOutboundHandler() (*httpclient.OutboundHandler, error) {
	h, err := httpclient.NewOutboundHandler(
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
func NewHttpInboundHandler() (*httpserver.InboundHandler, error) {
	h, err := httpserver.NewInboundHandler(
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
