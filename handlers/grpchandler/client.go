package grpchandler

import (
	"github.com/StanzaSystems/sdk-go/handlers"
)

type OutboundHandler struct {
	*handlers.OutboundHandler
}

// NewOutboundHandler returns a new OutboundHandler
func NewOutboundHandler() (*OutboundHandler, error) {
	h, err := handlers.NewOutboundHandler()
	if err != nil {
		return nil, err
	}
	return &OutboundHandler{h}, nil
}
