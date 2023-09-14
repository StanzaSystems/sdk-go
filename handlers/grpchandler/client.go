package grpchandler

import (
	"context"
	"net/http"

	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/keys"
	"github.com/StanzaSystems/sdk-go/otel"

	otel_codes "go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

// NewUnaryClientInterceptor returns a Guarded grpc.UnaryClientInterceptor
func (h *OutboundHandler) NewUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption,
	) error {
		ctx, span := h.start(ctx, method)
		defer span.End()

		if guard := h.Guard(ctx, span, nil); guard.Blocked() {
			return h.blocked(span, guard)
		} else {
			err := invoker(h.headers(ctx, guard.Token()), method, req, reply, cc, callOpts...)
			return h.allowed(span, guard, err)
		}
	}
}

// NewStreamClientInterceptor returns a Guarded grpc.StreamClientInterceptor
func (h *OutboundHandler) NewStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		callOpts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx, span := h.start(ctx, method)
		defer span.End()

		if guard := h.Guard(ctx, span, nil); guard.Blocked() {
			return nil, h.blocked(span, guard)
		} else {
			s, err := streamer(h.headers(ctx, guard.Token()), desc, cc, method, callOpts...)
			return s, h.allowed(span, guard, err)
		}
	}
}

func (h *OutboundHandler) headers(ctx context.Context, token string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	md = md.Copy()
	if token != "" {
		md.Set("X-Stanza-Token", token)
	}
	if len(md.Get("User-Agent")) == 0 {
		md.Set("User-Agent", global.UserAgent())
	}
	if ctx.Value(keys.OutboundHeadersKey) != nil {
		for k, v := range ctx.Value(keys.OutboundHeadersKey).(http.Header) {
			md.Set(k, v...)
		}
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func (h *OutboundHandler) start(ctx context.Context, spanName string) (context.Context, trace.Span) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	h.Propagator().Inject(ctx, &otel.MetadataSupplier{Metadata: &md})
	ctx = metadata.NewOutgoingContext(ctx, md)

	ctx, span := h.Tracer().Start(
		ctx,
		spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.RPCSystemGRPC),
	)

	return ctx, span
}

func (h *OutboundHandler) allowed(span trace.Span, guard *handlers.Guard, err error) error {
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

func (h *OutboundHandler) blocked(span trace.Span, guard *handlers.Guard) error {
	// TODO: codes.ResourceExhausted is wrong for failed token check
	span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(codes.ResourceExhausted)))
	span.SetStatus(otel_codes.Error, guard.BlockMessage())
	return status.Errorf(codes.ResourceExhausted, guard.BlockMessage())
}
