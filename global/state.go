package global

import (
	"context"
	"sync"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.9.0"
)

type state struct {
	// application details set by SDK clients
	name        string // defaults to "default_application"
	release     string // defaults to "0.0.0"
	environment string // defaults to "dev"

	// stanza
	stanzaHub string // defaults to "localhost:9510"

	// sentinel
	ds        datasource.DataSource
	resources []string

	// otel
	otelResource *resource.Resource
}

var (
	globalState = state{
		name:        "default_application",
		release:     "0.0.0",
		environment: "dev",
		stanzaHub:   "localhost:9510",
	}
	globalStateLock = &sync.RWMutex{}
	initOnce        sync.Once
)

func Name() string {
	return globalState.name
}

func Release() string {
	return globalState.release
}

func Environment() string {
	return globalState.environment
}

func NewState(name, rel, env, hub string) {
	initOnce.Do(func() {
		// prepare for global state mutation
		globalStateLock.Lock()
		defer globalStateLock.Unlock()

		// initialize new global state
		globalState = state{
			name:        name,
			release:     rel,
			environment: env,
			stanzaHub:   hub,
		}

		// connect to stanzahub?
		// -- datasource for sentinel but where do we otlp otel metrics/traces?
		// -- do we need to "register" name/ver/env?
	})
}

func NewResource(resName string) error {
	globalStateLock.Lock()
	defer globalStateLock.Unlock()

	globalState.resources = append(globalState.resources, resName)
	return nil
}

func GetDataSource() datasource.DataSource {
	return globalState.ds
}

func SetDataSource(ds datasource.DataSource) error {
	globalStateLock.Lock()
	defer globalStateLock.Unlock()

	globalState.ds = ds
	return nil
}

func GetOtelResource() *resource.Resource {
	return globalState.otelResource
}

func SetOtelResource(ctx context.Context) error {
	globalStateLock.Lock()
	defer globalStateLock.Unlock()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(globalState.name),
			semconv.ServiceVersionKey.String(globalState.release),
			semconv.DeploymentEnvironmentKey.String(globalState.environment),
		),
	)
	if err != nil {
		return err
	}
	globalState.otelResource = res
	return nil
}
