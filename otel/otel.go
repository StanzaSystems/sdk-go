package otel

import (
	"context"
	"os"

	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	ot "go.opentelemetry.io/otel/trace"
)

var (
	mp  *metric.MeterProvider
	tp  *trace.TracerProvider
	res *resource.Resource
	err error
)

func Init(ctx context.Context, name, rel, env string) (func(), error) {
	res, err = resource.New(ctx,
		resource.WithHost(),
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(name),
			semconv.ServiceVersionKey.String(rel),
			semconv.DeploymentEnvironmentKey.String(env),
		),
	)

	mp = &metric.MeterProvider{}
	tp = &trace.TracerProvider{}
	return func() {
		mp.Shutdown(ctx)
		tp.Shutdown(ctx)
		logging.Debug("gracefully shutdown opentelemetry exporter")
	}, err
}

func InitMetricProvider(ctx context.Context, mc *hubv1.MetricConfig, token string) error {
	if os.Getenv("STANZA_DEBUG") != "" || os.Getenv("STANZA_OTEL_DEBUG") != "" {
		if mp, err = initDebugMeter(res); err != nil {
			panic(err)
		}
	} else {
		if mp, err = initGrpcMeter(ctx, res, mc, token); err != nil {
			return err
		}
	}
	return nil
}

func InitTraceProvider(ctx context.Context, tc *hubv1.TraceConfig, token string) error {
	if os.Getenv("STANZA_DEBUG") != "" || os.Getenv("STANZA_OTEL_DEBUG") != "" {
		if tp, err = initDebugTracer(res, tc); err != nil {
			panic(err)
		}
	} else {
		if tp, err = initGrpcTracer(ctx, res, tc, token); err != nil {
			return err
		}
	}
	return nil
}

// GetTextMapPropagator is a passthrough helper function
func GetTextMapPropagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

// GetTracePropagator is a passthrough helper function
func GetTracerProvider() ot.TracerProvider {
	return otel.GetTracerProvider()
}
