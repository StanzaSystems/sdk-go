package stanza

import (
	httphandler "github.com/StanzaSystems/sdk-go/handlers/http"
)

func NewHttpInboundHandler(decorator string) (*httphandler.InboundHandler, error) {
	return httphandler.NewInboundHandler(gs.client.Name, decorator, SentinelEnabled())
}
