package otel

import (
	"context"

	"github.com/StanzaSystems/sdk-go/global"

	otelotel "go.opentelemetry.io/otel"
	otelglobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.9.0"
)

func Init(ctx context.Context) error {
	// TODO: connect to an otel collector here?
	// TODO: connect to stanza-hub and get an otel config?

	if err := global.SetOtelConfig(
		otelglobal.MeterProvider(),
		otelotel.GetTextMapPropagator()); err != nil {
		return err
	}
	return nil
}

// Add additional resource attributes via the OTEL_RESOURCE_ATTRIBUTES environment variable
// https://opentelemetry.io/docs/concepts/sdk-configuration/general-sdk-configuration/#otel_resource_attributes
func Resource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(global.Name()),
			semconv.ServiceVersionKey.String(global.Release()),
			semconv.DeploymentEnvironmentKey.String(global.Environment()),
		),
	)
}
