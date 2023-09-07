package stanza

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/handlers/grpchandler"
	"github.com/StanzaSystems/sdk-go/keys"
	"github.com/StanzaSystems/sdk-go/logging"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type GuardOpt struct {
	Feature       string
	PriorityBoost int32
	DefaultWeight float32
	MethodFilter  []string
}

type GuardRequest struct {
	Context context.Context
	Headers http.Header
	Name    string
	URL     string
	Body    io.Reader
	Opt     *GuardOpt // overrides OTEL baggage (if exists)
}

// HttpGet is a helper function to Guard an outbound HTTP GET
func HttpGet(req GuardRequest) (*http.Response, error) {
	h, err := NewHttpOutboundHandler()
	if err != nil {
		logging.Error(fmt.Errorf("failed to create HTTP outbound handler: %v", err))
		return nil, err
	}
	return h.Get(ctx(req), req.Name, req.URL)
}

// HttpPost is a helper function to Guard an outbound HTTP POST
func HttpPost(req GuardRequest) (*http.Response, error) {
	h, err := NewHttpOutboundHandler()
	if err != nil {
		logging.Error(fmt.Errorf("failed to create HTTP outbound handler: %v", err))
		return nil, err
	}
	return h.Post(ctx(req), req.Name, req.URL, req.Body)
}

// UnaryServerInterceptor is a helper function to Guard an inbound grpc unary server
func UnaryServerInterceptor(guardName string, opts ...GuardOpt) grpc.UnaryServerInterceptor {
	h, err := grpchandler.NewInboundHandler()
	if err != nil {
		logging.Error(err)
		return nil
	}
	methodFilter, featureName, boost, weight := opt(opts)
	return h.NewUnaryServerInterceptor(methodFilter, guardName, featureName, boost, weight)
}

// StreamServerInterceptor is a helper function to Guard an inbound grpc streaming server
func StreamServerInterceptor(guardName string, opts ...GuardOpt) grpc.StreamServerInterceptor {
	h, err := grpchandler.NewInboundHandler()
	if err != nil {
		logging.Error(err)
		return nil
	}
	methodFilter, featureName, boost, weight := opt(opts)
	return h.NewStreamServerInterceptor(methodFilter, guardName, featureName, boost, weight)
}

// Guard is a helper function to Guard any arbitrary block of code
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

func opt(opts []GuardOpt) ([]string, string, int32, float32) {
	mf := []string{}
	fn := ""
	pb := int32(0)
	dw := float32(0)
	if len(opts) == 1 {
		mf = opts[0].MethodFilter
		fn = opts[0].Feature
		pb = opts[0].PriorityBoost
		dw = opts[0].DefaultWeight
	}
	return mf, fn, pb, dw
}

func ctx(req GuardRequest) context.Context {
	if req.Context == nil {
		req.Context = context.Background()
	}
	if req.Headers != nil {
		req.Context = context.WithValue(req.Context, keys.OutboundHeadersKey, req.Headers)
	} else {
		req.Context = context.WithValue(req.Context, keys.OutboundHeadersKey, make(http.Header))
	}
	if req.Opt != nil {
		if req.Opt.Feature != "" {
			req.Context = context.WithValue(req.Context, keys.StanzaFeatureNameKey, req.Opt.Feature)
		}
		if req.Opt.PriorityBoost != 0 {
			req.Context = context.WithValue(req.Context, keys.StanzaPriorityBoostKey, req.Opt.PriorityBoost)
		}
		if req.Opt.DefaultWeight != 0 {
			req.Context = context.WithValue(req.Context, keys.StanzaDefaultWeightKey, req.Opt.DefaultWeight)
		}
	}
	return req.Context
}
