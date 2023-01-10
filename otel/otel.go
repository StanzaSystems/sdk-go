package otel

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/StanzaSystems/sdk-go/logging"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var config = Config{
	traceRatio: 0.001, // Percentage of traces to sample (default: 0.001)
}

func Init(ctx context.Context, name, rel, env string) error {
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
		return fmt.Errorf("creating opentelemetry resource: %w", err)
	}

	if os.Getenv("STANZA_DEFAULT_TRACE_RATIO") != "" {
		newRatio, err := strconv.ParseFloat(os.Getenv("STANZA_DEFAULT_TRACE_RATIO"), 32)
		if err != nil {
			if err := SetTraceRatio(newRatio); err != nil {
				logging.Error(err)
			}
		}
	}

	if os.Getenv("STANZA_DEBUG") != "" {
		if _, err := initDebugMeter(res); err != nil {
			panic(err)
		}
		if _, err := initDebugTracer(res); err != nil {
			panic(err)
		}
	} else {
		if _, err := initGrpcMeter(ctx, res); err != nil {
			panic(err) // TODO: don't panic here
		}
		if _, err := initGrpcTracer(ctx, res); err != nil {
			panic(err) // TODO: don't panic here
		}
	}
	// Handle shutdown to ensure all sub processes are closed correctly and telemetry is exported
	//
	// TODO: add something like the below (but NOT just deferred from here)
	// defer func() {
	// 	_ = exp.Shutdown(ctx)
	// 	_ = tp.Shutdown(ctx)
	// }()
	return nil
}

func SetTraceRatio(r float64) error {
	if r <= 1.0 && r >= 0.0 {
		config.traceRatio = r
		return nil
	} else {
		return fmt.Errorf("invalid trace ratio: %v", r)
	}
}
