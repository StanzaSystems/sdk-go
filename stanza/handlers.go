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
func NewHttpInboundHandler() (*httphandler.InboundHandler, error) {
	return httphandler.NewInboundHandler()
}

// gRPC Client
func NewGrpcOutboundHandler() (*grpchandler.OutboundHandler, error) {
	return grpchandler.NewOutboundHandler()
}

// gRPC Server
func NewGrpcInboundHandler() (*grpchandler.InboundHandler, error) {
	return grpchandler.NewInboundHandler()
}
