package otel

import (
	"context"
	"errors"

	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

type SetupConfig struct {
	ServiceName        string
	ServiceVersion     string
	ServiceEnvironment string
	Headers            map[string]string
	MetricConfig       *hubv1.MetricConfig
	TraceConfig        *hubv1.TraceConfig
}

// Setup bootstraps the OpenTelemetry export pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func Setup(ctx context.Context, sc SetupConfig) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Setup resource.
	res, err := newResource(ctx, sc.ServiceName, sc.ServiceVersion, sc.ServiceEnvironment)
	if err != nil {
		handleErr(err)
		return
	}

	// Setup trace provider.
	tracerProvider, err := newTraceProvider(ctx, res, sc.Headers, sc.TraceConfig.GetCollectorUrl(), float64(sc.TraceConfig.GetSampleRateDefault()))
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Setup meter provider.
	meterProvider, err := newMeterProvider(ctx, res, sc.Headers, sc.MetricConfig.GetCollectorUrl())
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return
}

func newResource(ctx context.Context, serviceName, serviceVersion, serviceEnvironment string) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithHost(),
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.DeploymentEnvironmentKey.String(serviceEnvironment),
		),
	)
}
