package stanza

import (
	"github.com/StanzaSystems/sdk-go/handlers/grpchandler"
	"github.com/StanzaSystems/sdk-go/handlers/httphandler"
)

// HTTP Client
func NewHttpOutboundHandler(gn string, fn *string, pb *int32, dw *float32, kv *map[string]string) (*httphandler.OutboundHandler, error) {
	return httphandler.NewOutboundHandler(gn, fn, pb, dw, kv)
}

// HTTP Server
func NewHttpInboundHandler(gn string, fn *string, pb *int32, dw *float32, kv *map[string]string) (*httphandler.InboundHandler, error) {
	return httphandler.NewInboundHandler(gn, fn, pb, dw, kv)
}

// gRPC Client
func NewGrpcOutboundHandler(gn string, fn *string, pb *int32, dw *float32, kv *map[string]string) (*grpchandler.OutboundHandler, error) {
	return grpchandler.NewOutboundHandler(gn, fn, pb, dw, kv)
}

// gRPC Server
func NewGrpcInboundHandler(gn string, fn *string, pb *int32, dw *float32, kv *map[string]string) (*grpchandler.InboundHandler, error) {
	return grpchandler.NewInboundHandler(gn, fn, pb, dw, kv)
}
