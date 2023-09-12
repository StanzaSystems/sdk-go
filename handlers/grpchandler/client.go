package grpchandler

import (
	"github.com/StanzaSystems/sdk-go/handlers"
)

type OutboundHandler struct {
	*handlers.OutboundHandler
}

// NewOutboundHandler returns a new OutboundHandler
func NewOutboundHandler(gn string, fn *string, pb *int32, dw *float32) (*OutboundHandler, error) {
	h, err := handlers.NewOutboundHandler(gn, fn, pb, dw)
	if err != nil {
		return nil, err
	}
	return &OutboundHandler{h}, nil
}
