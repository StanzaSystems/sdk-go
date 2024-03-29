package grpchandler

import (
	"context"

	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/otel"

	otel_codes "go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type InboundHandler struct {
	*handlers.InboundHandler
}

// NewInboundHandler returns a new InboundHandler
func NewInboundHandler(gn string, fn *string, pb *int32, dw *float32, kv *map[string]string) (*InboundHandler, error) {
	h, err := handlers.NewInboundHandler(gn, fn, pb, dw, kv)
	if err != nil {
		return nil, err
	}

	return &InboundHandler{h}, nil
}

// NewUnaryServerInterceptor returns a Guarded grpc.UnaryServerInterceptor
func (h *InboundHandler) NewUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		ctx, span, tokens := h.start(ctx, info.FullMethod)
		defer span.End()

		if guard := h.Guard(ctx, span, tokens); guard.Blocked() {
			return nil, h.blocked(span, guard)
		} else {
			next, err := handler(ctx, req)
			return next, h.allowed(span, guard, err)
		}
	}
}

// NewStreamServerInterceptor returns a Guarded grpc.StreamServerInterceptor
func (h *InboundHandler) NewStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx, span, tokens := h.start(stream.Context(), info.FullMethod)
		defer span.End()

		if guard := h.Guard(ctx, span, tokens); guard.Blocked() {
			return h.blocked(span, guard)
		} else {
			err := handler(srv, stream)
			return h.allowed(span, guard, err)
		}
	}
}

func (h *InboundHandler) start(ctx context.Context, spanName string) (context.Context, trace.Span, []string) {
	tokens := []string{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	} else {
		tokens = md.Get("x-stanza-token")
	}
	ctx = h.Propagator().Extract(ctx, &otel.MetadataSupplier{Metadata: &md})

	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(semconv.RPCSystemGRPC),
	}
	if s := trace.SpanContextFromContext(ctx); s.IsValid() && s.IsRemote() {
		opts = append(opts, trace.WithLinks(trace.Link{SpanContext: s}))
	}

	ctx, span := h.Tracer().Start(ctx, spanName, opts...)
	return ctx, span, tokens
}

func (h *InboundHandler) allowed(span trace.Span, guard *handlers.Guard, err error) error {
	s, _ := status.FromError(err)
	span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(s.Code())))

	if err != nil {
		span.SetStatus(otel_codes.Error, s.Message())
		span.RecordError(err)
		guard.End(guard.Failure)
	} else {
		span.SetStatus(otel_codes.Ok, "OK")
		guard.End(guard.Success)
	}
	return err
}

func (h *InboundHandler) blocked(span trace.Span, guard *handlers.Guard) error {
	// TODO: codes.ResourceExhausted is wrong for failed token check
	span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(codes.ResourceExhausted)))
	span.SetStatus(otel_codes.Error, guard.BlockMessage())
	return status.Errorf(codes.ResourceExhausted, guard.BlockMessage())
}
