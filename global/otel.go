package global

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
)

type otel struct {
	MeterProvider metric.MeterProvider
	Propagators   propagation.TextMapPropagator
}

func GetOtelConfig() *otel {
	return gs.otel
}

func SetOtelConfig(mp metric.MeterProvider, p propagation.TextMapPropagator) error {
	gsLock.Lock()
	defer gsLock.Unlock()

	gs.otel.MeterProvider = mp
	gs.otel.Propagators = p
	return nil
}
