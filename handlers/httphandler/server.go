package httphandler

import (
	"github.com/StanzaSystems/sdk-go/handlers"
)

type InboundHandler struct {
	*handlers.InboundHandler
}

// NewInboundHandler returns a new InboundHandler
func NewInboundHandler() (*InboundHandler, error) {
	h, err := handlers.NewInboundHandler()
	if err != nil {
		return nil, err
	}
	return &InboundHandler{h}, nil
}

// TODO: Add generic HTTP server middleware
