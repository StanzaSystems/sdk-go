package stanza

import (
	"context"
	"crypto/tls"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	httphandler "github.com/StanzaSystems/sdk-go/handlers/http"
	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type state struct {
	clientOpt       *ClientOptions
	clientId        uuid.UUID
	hubConn         *grpc.ClientConn
	hubAuthClient   hubv1.AuthServiceClient
	hubConfigClient hubv1.ConfigServiceClient
	hubQuotaClient  hubv1.QuotaServiceClient
	inboundHandlers []*httphandler.InboundHandler

	// stored from GetBearerToken request
	bearerToken     string
	bearerTokenTime time.Time

	// stored from GetServiceConfig polling
	svcConfig        *hubv1.ServiceConfig
	svcConfigTime    time.Time
	svcConfigVersion string

	// stored from GetDecoratorConfig polling
	decoratorConfig        map[string]*hubv1.DecoratorConfig
	decoratorConfigTime    map[string]time.Time
	decoratorConfigVersion map[string]string

	// OTEL
	otelInit                    bool
	otelMetricProviderConnected bool
	otelTraceProviderConnected  bool

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
			clientOpt:                   &co,
			clientId:                    uuid.New(),
			hubConn:                     nil,
			bearerToken:                 "",
			bearerTokenTime:             time.Time{},
			svcConfig:                   &hubv1.ServiceConfig{},
			svcConfigTime:               time.Time{},
			svcConfigVersion:            "",
			decoratorConfig:             map[string]*hubv1.DecoratorConfig{},
			decoratorConfigTime:         map[string]time.Time{},
			decoratorConfigVersion:      map[string]string{},
			otelInit:                    false,
			otelMetricProviderConnected: false,
			otelTraceProviderConnected:  false,
			sentinelConnected:           false,
			sentinelConnectedTime:       time.Time{},
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
			if gs.hubConn != nil {
				if gs.hubConn.GetState() == connectivity.Ready {
					otelShutdown = OtelStartup(ctx)
					GetServiceConfig(ctx)
					GetDecoratorConfigs(ctx)
					for _, ih := range gs.inboundHandlers {
						ih.SetQuotaServiceClient(gs.hubQuotaClient)
					}
				} else {
					gs.hubConn.Connect()
				}
			} else {
				creds := credentials.NewTLS(&tls.Config{})
				if os.Getenv("STANZA_HUB_NO_TLS") != "" { // disable TLS for local Hub development
					creds = insecure.NewCredentials()
				}
				opts := []grpc.DialOption{
					grpc.WithTransportCredentials(creds),
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
					gs.hubQuotaClient = hubv1.NewQuotaServiceClient(hubConn)
				}
			}
		}
	}
}
