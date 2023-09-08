package stanza

import (
	"github.com/StanzaSystems/sdk-go/handlers/grpchandler"
	"github.com/StanzaSystems/sdk-go/handlers/httphandler"
)

// HTTP Client
func NewHttpOutboundHandler() (*httphandler.OutboundHandler, error) {
	return httphandler.NewOutboundHandler()
}

// HTTP Server
func NewHttpInboundHandler(gn string, fn *string, pb *int32, dw *float32) (*httphandler.InboundHandler, error) {
	return httphandler.NewInboundHandler(gn, fn, pb, dw)
}

// gRPC Client
func NewGrpcOutboundHandler() (*grpchandler.OutboundHandler, error) {
	return grpchandler.NewOutboundHandler()
}

// gRPC Server
func NewGrpcInboundHandler(gn string, fn *string, pb *int32, dw *float32) (*grpchandler.InboundHandler, error) {
	return grpchandler.NewInboundHandler(gn, fn, pb, dw)
}
