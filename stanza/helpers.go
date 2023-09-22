package stanza

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/handlers/grpchandler"
	"github.com/StanzaSystems/sdk-go/handlers/httphandler"
	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/StanzaSystems/sdk-go/otel"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type GuardOpt struct {
	Feature       *string
	PriorityBoost *int32
	DefaultWeight *float32
}

// HttpServer is a helper function to Guard inbound HTTP requests
func HttpServer(guardName string, opts ...GuardOpt) (*httphandler.InboundHandler, error) {
	return NewHttpInboundHandler(withOpts(guardName, opts...))
}

func GuardMiddleware(next func(w http.ResponseWriter, r *http.Request), guardName string, opts ...GuardOpt) func(w http.ResponseWriter, r *http.Request) {
	h, err := NewHttpInboundHandler(withOpts(guardName, opts...))
	if err != nil {
		logging.Error(fmt.Errorf("no HTTP inbound handler, failing open"))
		if h != nil {
			h.FailOpen(context.Background())
		}
		return next
	}
	return h.GuardHandlerFunction(next)
}

func GuardHandler(next http.Handler, guardName string, opts ...GuardOpt) http.Handler {
	h, err := NewHttpInboundHandler(withOpts(guardName, opts...))
	if err != nil {
		logging.Error(fmt.Errorf("no HTTP inbound handler, failing open"))
		if h != nil {
			h.FailOpen(context.Background())
		}
		return next
	}
	return h.GuardHandler(next)
}

// HttpGet is a helper function to Guard an outbound HTTP GET
func HttpGet(ctx context.Context, guardName, url string, opts ...GuardOpt) (*http.Response, error) {
	h, err := NewHttpOutboundHandler(withOpts(guardName, opts...))
	if err != nil {
		logging.Error(fmt.Errorf("failed to create HTTP outbound handler: %v", err))
		return nil, err
	}
	return h.Get(ctx, url)
}

// HttpPost is a helper function to Guard an outbound HTTP POST
func HttpPost(ctx context.Context, guardName, url string, body io.Reader, opts ...GuardOpt) (*http.Response, error) {
	h, err := NewHttpOutboundHandler(withOpts(guardName, opts...))
	if err != nil {
		logging.Error(fmt.Errorf("failed to create HTTP outbound handler: %v", err))
		return nil, err
	}
	return h.Post(ctx, url, body)
}

// UnaryServerInterceptor is a helper function to Guard an inbound grpc unary server
func UnaryServerInterceptor(guardName string, opts ...GuardOpt) grpc.UnaryServerInterceptor {
	h, err := grpchandler.NewInboundHandler(withOpts(guardName, opts...))
	if err != nil {
		logging.Error(err)
		return nil
	}
	return h.NewUnaryServerInterceptor()
}

// StreamServerInterceptor is a helper function to Guard an inbound grpc streaming server
func StreamServerInterceptor(guardName string, opts ...GuardOpt) grpc.StreamServerInterceptor {
	h, err := grpchandler.NewInboundHandler(withOpts(guardName, opts...))
	if err != nil {
		logging.Error(err)
		return nil
	}
	return h.NewStreamServerInterceptor()
}

// UnaryClientInterceptor is a helper function to Guard an outbound grpc unary client
func UnaryClientInterceptor(guardName string, opts ...GuardOpt) grpc.UnaryClientInterceptor {
	h, err := grpchandler.NewOutboundHandler(withOpts(guardName, opts...))
	if err != nil {
		logging.Error(err)
		return nil
	}
	return h.NewUnaryClientInterceptor()
}

// StreamClientInterceptor is a helper function to Guard an outbound grpc streaming client
func StreamClientInterceptor(guardName string, opts ...GuardOpt) grpc.StreamClientInterceptor {
	h, err := grpchandler.NewOutboundHandler(withOpts(guardName, opts...))
	if err != nil {
		logging.Error(err)
		return nil
	}
	return h.NewStreamClientInterceptor()
}

// Guard is a helper function to Guard any arbitrary block of code
func Guard(ctx context.Context, guardName string, opts ...GuardOpt) *handlers.Guard {
	h, err := handlers.NewHandler(withOpts(guardName, opts...))
	if err != nil {
		err = fmt.Errorf("failed to create guard handler: %s", err)
		logging.Error(err)
		return h.NewGuard(ctx, nil, nil, err)
	}
	traceOpts := []trace.SpanStartOption{
		// WithAttributes?
		trace.WithSpanKind(trace.SpanKindInternal),
	}
	ctx, span := h.Tracer().Start(ctx, guardName, traceOpts...)
	defer span.End()
	return h.Guard(ctx, span, nil)
}

// ContextWithHeaders is a helper function which extracts and OTEL TraceContext, Baggage,
// and StanzaHeaders from a given http.Request into a context.Context.
func ContextWithHeaders(r *http.Request) context.Context {
	return otel.ContextWithHeaders(r)
}

func withOpts(gn string, opts ...GuardOpt) (string, *string, *int32, *float32) {
	var fn *string
	var pb *int32
	var dw *float32
	if len(opts) == 1 {
		if opts[0].Feature != nil {
			fn = opts[0].Feature
		}
		if opts[0].PriorityBoost != nil {
			pb = opts[0].PriorityBoost
		}
		if opts[0].DefaultWeight != nil {
			dw = opts[0].DefaultWeight
		}
	}
	return gn, fn, pb, dw
}
