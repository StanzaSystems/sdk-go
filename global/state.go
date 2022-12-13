package global

import (
	"sync"

	"github.com/alibaba/sentinel-golang/ext/datasource"
)

type state struct {
	// base required options
	appName     string
	environment string
	resources   []string

	// do we need this? probably store an otlp connection instead
	stanzaHub string

	// sentinel datasource
	ds datasource.DataSource
}

var (
	globalState     = state{}
	globalStateLock = &sync.RWMutex{}
	initOnce        sync.Once
)

func AppName() string {
	return globalState.appName
}

func NewState(app, env, hub string) {
	initOnce.Do(func() {
		// prepare for global state mutation
		globalStateLock.Lock()
		defer globalStateLock.Unlock()

		// initialize new global state
		globalState = state{
			appName:     app,
			environment: env,
			stanzaHub:   hub,
		}

		// connect to stanzahub?
		// -- datasource for sentinel but where do we otlp otel metrics/traces?
		// -- do we need to "register" AppName? (so the GUI can offer makign configs for it)
	})
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

func NewResource(resName string) error {
	globalStateLock.Lock()
	defer globalStateLock.Unlock()

	globalState.resources = append(globalState.resources, resName)
	return nil
}
