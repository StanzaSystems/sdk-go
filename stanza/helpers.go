package stanza

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/handlers/grpchandler"
	"github.com/StanzaSystems/sdk-go/logging"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// Optional arguments
type GuardOpt struct {
	Headers       http.Header
	Feature       string
	PriorityBoost int32
	DefaultWeight float32
}

type GuardRequest struct {
	Context context.Context
	Name    string
	URL     string
	Body    io.Reader
}

// HTTP Outbound Helper: GET
func HttpGet(req GuardRequest) (*http.Response, error) {
	if req.Context == nil {
		req.Context = context.Background()
	}
	h, err := NewHttpOutboundHandler()
	if err != nil {
		logging.Error(fmt.Errorf("failed to create HTTP outbound handler: %v", err))
		return nil, err
	}
	return h.Get(req.Context, req.Name, req.URL)
}

// HTTP Outbound Helper: POST
func HttpPost(req GuardRequest) (*http.Response, error) {
	if req.Context == nil {
		req.Context = context.Background()
	}
	h, err := NewHttpOutboundHandler()
	if err != nil {
		logging.Error(fmt.Errorf("failed to create HTTP outbound handler: %v", err))
		return nil, err
	}
	return h.Post(req.Context, req.Name, req.URL, req.Body)
}

// gRPC Unary Server Interceptor
func UnaryServerInterceptor(guardName string) grpc.UnaryServerInterceptor {
	h, err := grpchandler.NewInboundHandler()
	if err != nil {
		logging.Error(err)
		return nil
	}
	return h.NewUnaryServerInterceptor(guardName)
}

// gRPC Stream Server Interceptor
func StreamServerInterceptor(guardName string) grpc.StreamServerInterceptor {
	h, err := grpchandler.NewInboundHandler()
	if err != nil {
		logging.Error(err)
		return nil
	}
	return h.NewStreamServerInterceptor(guardName)
}

// Generic Guard
func Guard(ctx context.Context, name string) *handlers.Guard {
	h, err := handlers.NewHandler()
	if err != nil {
		err = fmt.Errorf("failed to create guard handler: %s", err)
		logging.Error(err)
		return h.NewGuardError(ctx, nil, nil, err)
	}
	opts := []trace.SpanStartOption{
		// WithAttributes?
		trace.WithSpanKind(trace.SpanKindInternal),
	}
	ctx, span := h.Tracer().Start(ctx, name, opts...)
	defer span.End()
	return h.NewGuard(ctx, span, name, []string{})
}
