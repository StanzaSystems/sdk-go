package stanza

import (
	"github.com/StanzaSystems/sdk-go/handlers/httphandler"
)

// HTTP Client
func NewHttpOutboundHandler() (*httphandler.OutboundHandler, error) {
	return httphandler.NewOutboundHandler()
}

// HTTP Server
func NewHttpInboundHandler() (*httphandler.InboundHandler, error) {
	return httphandler.NewInboundHandler()
}
