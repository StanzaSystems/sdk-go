package grpchandler

import (
	"context"

	"github.com/StanzaSystems/sdk-go/handlers"

	otel_codes "go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (h *InboundHandler) NewUnaryServerInterceptor(guardName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// TODO: otel baggage propagation through via grpc
		// ctx = h.Propagator().Extract(ctx, propagation.HeaderCarrier(req.Header))

		// requestMetadata, _ := metadata.FromIncomingContext(ctx)
		// metadataCopy := requestMetadata.Copy()

		// entries, spanCtx := Extract(ctx, &metadataCopy, opts...)
		// ctx = baggage.C
		// name, attr := spanInfo(info.FullMethod, peerFromCtx(ctx))

		// TODO: extract more attributes from an existing (parent) span?
		ctx, span := h.Tracer().Start(
			ctx,
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(semconv.RPCSystemGRPC),
		)
		// ctx, span := h.Tracer().Start(ctx, guardName, traceOpts...)
		defer span.End()

		// guard := h.NewGuard(addKeys(ctx, opts...), span, guardName, req.Header.Values("x-stanza-token"))
		// TODO: where to check for X-Stanza-Token?
		tokens := []string{}
		guard := h.NewGuard(ctx, span, guardName, tokens)

		// Stanza Blocked
		if guard.Blocked() {
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(codes.ResourceExhausted)))
			return nil, status.Errorf(codes.ResourceExhausted, guard.BlockMessage())
		}

		// Stanza Allowed
		next, err := handler(ctx, req) // intercept following middleware for guard.End() status
		if err != nil {
			s, ok := status.FromError(err)
			if ok {
				span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(s.Code())))
			} else {
				span.SetStatus(otel_codes.Error, err.Error())
			}
			span.RecordError(err)
			guard.End(guard.Failure)
		} else {
			span.SetStatus(otel_codes.Ok, "OK")
			guard.End(guard.Success)
		}
		return next, err
	}
}

func (h *InboundHandler) NewStreamServerInterceptor(guardName string) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// TODO: otel baggage propagation through via grpc
		// ctx = h.Propagator().Extract(ctx, propagation.HeaderCarrier(req.Header))
		// ctx := h.Propagator().Extract(stream.Context(), nil)

		// TODO: extract more attributes from an existing (parent) span?
		ctx, span := h.Tracer().Start(
			stream.Context(),
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(semconv.RPCSystemGRPC),
		)
		defer span.End()

		// guard := h.NewGuard(addKeys(ctx, opts...), span, guardName, req.Header.Values("x-stanza-token"))
		// TODO: where to check for X-Stanza-Token?
		tokens := []string{}
		guard := h.NewGuard(ctx, span, guardName, tokens)

		// Stanza Blocked
		if guard.Blocked() {
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(codes.ResourceExhausted)))
			return status.Errorf(codes.ResourceExhausted, guard.BlockMessage())
		}

		// Stanza Allowed
		err := handler(srv, stream) // intercept following middleware for guard.End() status
		if err != nil {
			s, ok := status.FromError(err)
			if ok {
				span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(s.Code())))
			} else {
				span.SetStatus(otel_codes.Error, err.Error())
			}
			span.RecordError(err)
			guard.End(guard.Failure)
		} else {
			span.SetStatus(otel_codes.Ok, "OK")
			guard.End(guard.Success)
		}
		return err
	}
}
