package global

import (
	"fmt"
	"sync"
	"time"

	"github.com/StanzaSystems/sdk-go/otel"
	"github.com/StanzaSystems/sdk-go/sentinel"
)

type resource struct {
	updated  time.Time
	sentinel sentinel.Config
	otel     otel.Config
}

type state struct {
	// base required options
	appName     string
	environment string
	stanzaHub   string

	// map of named resources
	resources map[string]resource

	// do we need to store an open connection handler?
	// could be grpc or otlp? -- or maybe both?
}

var (
	globalState     = state{}
	globalStateLock = &sync.RWMutex{}
)

func AppName() string {
	return globalState.appName
}

func NewState(app, env, hub string) error {
	if globalState.appName != "" ||
		globalState.stanzaHub != "" {
		return fmt.Errorf("already initialized global state")
	}

	// prepare for global state mutation
	globalStateLock.Lock()
	defer globalStateLock.Unlock()

	// initialize and set new global state
	globalState = state{
		appName:     app,
		environment: env,
		stanzaHub:   hub,
		resources:   make(map[string]resource),
	}

	// connect to stanzahub
	// -- establish grpc and/or oltp?
	// -- registering AppName

	return nil
}

func NewResource(resourceName string) error {
	if _, exists := globalState.resources[resourceName]; exists {
		return fmt.Errorf("duplicate resource named '%s' exists", resourceName)
	}

	// prepare for global state mutation
	globalStateLock.Lock()
	defer globalStateLock.Unlock()

	// initialize with default (empty) configs
	globalState.resources[resourceName] = resource{
		updated:  time.Now(),
		sentinel: sentinel.Config{},
		otel:     otel.Config{},
	}

	// TODO:
	// use goroutines (and waitgroup? with grpc streaming?) to add a listener
	// for getting globalConfig.AppName/resourceName configs (pull or streaming/pushed?)
	// something like:
	//   go fetchResourceConfig(resourceName)

	return nil
}
