package global

import (
	"sync"
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