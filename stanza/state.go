package stanza

import (
	"os"
	"sync"
)

type state struct {
	client *ClientOptions
}

var (
	gs       = state{}
	gsLock   = &sync.RWMutex{}
	initOnce sync.Once
)

func newState(client ClientOptions) {
	initOnce.Do(func() {
		// prepare for global state mutation
		gsLock.Lock()
		defer gsLock.Unlock()

		// initialize new global state
		gs = state{client: &client}

		// connect to stanzahub
		// -- store open gRPC conn
		// -- register name/ver/env
		// -- get otel config
	})
}

func SentinelEnabled() bool {
	if gs.client.DataSource != "" && os.Getenv("STANZA_NO_SENTINEL") == "" {
		return true
	}
	return false
}
