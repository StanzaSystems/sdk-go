package otel

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/StanzaSystems/sdk-go/logging"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var config = Config{
	traceRatio: 0.001, // Percentage of traces to sample (default: 0.001)
}

func Init(ctx context.Context, name, rel, env string) (func(), error) {
	res, err := resource.New(ctx,
		resource.WithHost(),
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(name),
			semconv.ServiceVersionKey.String(rel),
			semconv.DeploymentEnvironmentKey.String(env),
		),
	)
	if err != nil {
		return func() {}, fmt.Errorf("creating opentelemetry resource: %w", err)
	}

	if os.Getenv("STANZA_DEFAULT_TRACE_RATIO") != "" {
		newRatio, err := strconv.ParseFloat(os.Getenv("STANZA_DEFAULT_TRACE_RATIO"), 32)
		if err != nil {
			logging.Error(fmt.Errorf("parsing default trace ratio: %s", err))
		} else {
			if err := SetTraceRatio(newRatio); err != nil {
				logging.Error(err)
			}
		}
	}

	var mp *metric.MeterProvider
	var tp *trace.TracerProvider
	if os.Getenv("STANZA_DEBUG") != "" {
		if mp, err = initDebugMeter(res); err != nil {
			panic(err)
		}
		if tp, err = initDebugTracer(res); err != nil {
			panic(err)
		}
	} else {
		if mp, err = initGrpcMeter(ctx, res); err != nil {
			// TODO: don't panic here
			// but what should we do? Retry indefinitely?
			// (with exponential backoff to very infrequently?)
			panic(err)
		}
		if tp, err = initGrpcTracer(ctx, res); err != nil {
			panic(err) // TODO: don't panic here
		}
	}
	return func() {
		mp.Shutdown(ctx)
		tp.Shutdown(ctx)
	}, nil
}

func SetTraceRatio(r float64) error {
	if r <= 1.0 && r >= 0.0 {
		config.traceRatio = r
		return nil
	} else {
		return fmt.Errorf("invalid trace ratio: %v", r)
	}
}
