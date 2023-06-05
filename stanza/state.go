package stanza

import (
	"context"
	"crypto/tls"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	svcConfig        *hubv1.ServiceConfig
	svcConfigTime    time.Time
	svcConfigVersion string

	// OTEL
	otelConnected     bool
	otelConnectedTime time.Time

	// sentinel
	sentinelConnected     bool
	sentinelConnectedTime time.Time
	sentinelDatasource    string
}

var (
	gs       = state{}
	gsLock   = &sync.RWMutex{}
	initOnce sync.Once
)

func newState(ctx context.Context, co ClientOptions) func() {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	initOnce.Do(func() {
		// prepare for global state mutation
		gsLock.Lock()
		defer gsLock.Unlock()

		// initialize new global state
		gs = state{
			clientOpt:             &co,
			hubConn:               nil,
			bearerToken:           "",
			bearerTokenTime:       time.Time{},
			svcConfig:             nil,
			svcConfigTime:         time.Time{},
			svcConfigVersion:      "",
			otelConnected:         false,
			otelConnectedTime:     time.Time{},
			sentinelConnected:     false,
			sentinelConnectedTime: time.Time{},
		}
		gs.sentinelDatasource, _ = os.MkdirTemp("", "sentinel")

		// connect to stanza-hub
		go connectHub(ctx)
	})
	return func() {
		stop()
		if gs.hubConn != nil {
			gs.hubConn.Close()
			logging.Debug("disconnected from stanza hub", "url", gs.clientOpt.StanzaHub)
		}
	}
}

func connectHub(ctx context.Context) {
	const MIN_POLLING_TIME = 5 * time.Second

	otelShutdown := func() {}
	sentinelShutdown := func() {}
	for {
		select {
		case <-ctx.Done():
			otelShutdown()
			sentinelShutdown()
			return
		case <-time.After(MIN_POLLING_TIME):
			if gs.hubConn != nil { // AND some kind of healthcheck/ping on hubConn success?
				GetBearerToken(ctx)
				otelShutdown = OtelStartup(ctx)
				GetServiceConfig(ctx)
				sentinelShutdown = SentinelStartup(ctx)
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
