package global

import (
	"sync"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
)

type state struct {
	// application details set by SDK clients
	name        string // defaults to "unknown_service"
	release     string // defaults to "0.0.0"
	environment string // defaults to "dev"

	// stanza
	stanzaHub string // defaults to "localhost:9510"

	// sentinel
	sentinel *sentinel

	// otel
	otel *otel
}

type sentinel struct {
	ds        *datasource.DataSource
	resources []string
}

type otel struct {
	MeterProvider otelmetric.MeterProvider
	Propagators   propagation.TextMapPropagator
}

var (
	gs = state{
		name:        "unknown_service",
		release:     "0.0.0",
		environment: "dev",
		stanzaHub:   "localhost:9510",
	}
	gsLock   = &sync.RWMutex{}
	initOnce sync.Once
)

func Name() string {
	return gs.name
}

func Release() string {
	return gs.release
}

func Environment() string {
	return gs.environment
}

func NewState(name, rel, env, hub string) {
	initOnce.Do(func() {
		// prepare for global state mutation
		gsLock.Lock()
		defer gsLock.Unlock()

		// initialize new global state
		gs = state{
			name:        name,
			release:     rel,
			environment: env,
			stanzaHub:   hub,
			sentinel:    &sentinel{},
			otel:        &otel{},
		}

		// connect to stanzahub?
		// -- datasource for sentinel but where do we otlp otel metrics/traces?
		// -- do we need to "register" name/ver/env?
	})
}

func NewResource(resName string) error {
	gsLock.Lock()
	defer gsLock.Unlock()

	gs.sentinel.resources = append(gs.sentinel.resources, resName)
	return nil
}

func GetDataSource() *datasource.DataSource {
	return gs.sentinel.ds
}

func SetDataSource(ds datasource.DataSource) error {
	gsLock.Lock()
	defer gsLock.Unlock()

	gs.sentinel.ds = &ds
	return nil
}

func GetOtelConfig() *otel {
	return gs.otel
}

func SetOtelConfig(mp otelmetric.MeterProvider, p propagation.TextMapPropagator) error {
	gsLock.Lock()
	defer gsLock.Unlock()

	gs.otel.MeterProvider = mp
	gs.otel.Propagators = p
	return nil
}
