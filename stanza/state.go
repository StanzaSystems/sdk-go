package stanza

import (
	"context"
	"crypto/tls"
	"os"
	"sync"
	"time"

	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type state struct {
	clientOpt       *ClientOptions
	hubConn         *grpc.ClientConn
	hubAuthClient   hubv1.AuthServiceClient
	hubConfigClient hubv1.ConfigServiceClient

	// stored from GetBearerToken request
	bearerToken     string
	bearerTokenTime time.Time

	// stored from GetServiceConfig polling
	svcConfig        hubv1.ServiceConfig
	svcConfigTime    time.Time
	svcConfigVersion string
}

var (
	gs       = state{}
	gsLock   = &sync.RWMutex{}
	initOnce sync.Once
)

func newState(co ClientOptions) func() {
	ctx, ctxCancel := context.WithCancel(context.Background())
	initOnce.Do(func() {
		// prepare for global state mutation
		gsLock.Lock()
		defer gsLock.Unlock()

		// initialize new global state
		gs = state{
			clientOpt:        &co,
			hubConn:          nil,
			bearerToken:      "",
			svcConfigVersion: "",
		}

		// connect to stanza-hub
		go connectHub(ctx)
	})
	return func() {
		ctxCancel()
		if gs.hubConn != nil {
			gs.hubConn.Close()
		}
	}
}

func connectHub(ctx context.Context) {
	const MIN_POLLING_TIME = 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(MIN_POLLING_TIME):
			if gs.hubConn != nil {
				GetBearerToken(ctx)
				GetServiceConfig(ctx)
			} else {
				opts := []grpc.DialOption{
					grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
					// grpc.WithUserAgent(), // todo: SDK spec 
					// todo: add keepalives, backoff config, etc
				}
				hubConn, err := grpc.Dial(gs.clientOpt.StanzaHub, opts...)
				if err != nil {
					logging.Error(err,
						"msg", "failed to connect to stanza hub",
						"url", gs.clientOpt.StanzaHub)
				} else {
					logging.Debug(
						"connected to stanza hub",
						"url", gs.clientOpt.StanzaHub)
					gs.hubConn = hubConn
					gs.hubAuthClient = hubv1.NewAuthServiceClient(hubConn)
					gs.hubConfigClient = hubv1.NewConfigServiceClient(hubConn)
					hubConn.Connect()
				}
			}
		}
	}
}

func SentinelEnabled() bool {
	if gs.clientOpt.DataSource != "" && os.Getenv("STANZA_NO_SENTINEL") == "" {
		return true
	}
	return false
}
