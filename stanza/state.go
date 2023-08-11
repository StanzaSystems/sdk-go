package stanza

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/StanzaSystems/sdk-go/ca"
	"github.com/StanzaSystems/sdk-go/handlers/httphandler"
	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	MIN_POLLING_TIME = 15 * time.Second

	filePerms = 0660
)

type state struct {
	clientOpt       *ClientOptions
	clientId        uuid.UUID
	hubConn         *grpc.ClientConn
	hubAuthClient   hubv1.AuthServiceClient
	hubConfigClient hubv1.ConfigServiceClient
	hubQuotaClient  hubv1.QuotaServiceClient

	// HTTP
	httpInboundHandler  *httphandler.InboundHandler
	httpOutboundHandler *httphandler.OutboundHandler

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
	sentinelInit       bool
	sentinelInitTime   time.Time
	sentinelDatasource string
	sentinelRules      map[string]string
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
			sentinelInit:                false,
			sentinelInitTime:            time.Time{},
		}

		// pre-create empty sentinel rules files
		gs.sentinelDatasource, _ = os.MkdirTemp("", "sentinel")
		gs.sentinelRules = map[string]string{
			"circuitbreaker": filepath.Join(gs.sentinelDatasource, "circuitbreaker_rules.json"),
			"flow":           filepath.Join(gs.sentinelDatasource, "flow_rules.json"),
			"isolation":      filepath.Join(gs.sentinelDatasource, "isolation_rules.json"),
			"system":         filepath.Join(gs.sentinelDatasource, "system_rules.json"),
		}
		for _, fn := range gs.sentinelRules {
			err := os.WriteFile(fn, []byte("[]"), filePerms)
			if err != nil {
				logging.Error(err)
			}
		}

		if len(gs.clientOpt.Decorators) > 0 {
			for _, decorator := range gs.clientOpt.Decorators {
				gs.decoratorConfig[decorator] = &hubv1.DecoratorConfig{}
				gs.decoratorConfigTime[decorator] = time.Time{}
				gs.decoratorConfigVersion[decorator] = ""
			}
		}

		// end global state mutation
		gsLock.Unlock()

		// connect to stanza-hub
		if gs.hubConn == nil {
			hubConnect(ctx)
		}

		// start background polling for updates
		go hubPoller(ctx, MIN_POLLING_TIME)
	})
	return func() {
		stop()
		if gs.hubConn != nil {
			gs.hubConn.Close()
			logging.Debug("disconnected from stanza hub", "uri", gs.clientOpt.StanzaHub)
		}
	}
}

func hubConnect(ctx context.Context) (func(), func()) {
	tlsConfig := &tls.Config{}
	if caPath := os.Getenv("STANZA_AWS_ROOT_CA"); caPath != "" {
		tlsConfig.RootCAs = ca.AWSRootCAs(caPath)
	}
	creds := credentials.NewTLS(tlsConfig)
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
		gsLock.Lock()
		gs.hubConn = hubConn
		gs.hubAuthClient = hubv1.NewAuthServiceClient(hubConn)
		gs.hubConfigClient = hubv1.NewConfigServiceClient(hubConn)
		gs.hubQuotaClient = hubv1.NewQuotaServiceClient(hubConn)
		gsLock.Unlock()

		// attempt to establish hub connection (doesn't block)
		gs.hubConn.Connect()

		// block, waiting for up to 10 seconds for hub connection
		ctxWait, ctxWaitCancel := context.WithTimeout(ctx, 10*time.Second)
		defer ctxWaitCancel()
		gs.hubConn.WaitForStateChange(ctxWait, connectivity.Connecting)
		if gs.hubConn.GetState() == connectivity.Ready {
			logging.Info("connected to stanza hub", "uri", gs.clientOpt.StanzaHub)
			fetchConfigs(ctx, true)
			return OtelStartup(ctx), SentinelStartup(ctx)
		}
	}
	return func() {}, func() {}
}

func hubPoller(ctx context.Context, pollInterval time.Duration) {
	connectAttempt := 0
	otelShutdown := func() {}
	sentinelShutdown := func() {}
	for {
		select {
		case <-ctx.Done():
			otelShutdown()
			sentinelShutdown()
			return
		case <-time.After(pollInterval):
			if gs.hubConn != nil {
				if gs.hubConn.GetState() == connectivity.Ready {
					if connectAttempt > 0 {
						logging.Info(
							"connected to stanza hub",
							"uri", gs.clientOpt.StanzaHub,
							"attempt", connectAttempt,
						)
						connectAttempt = 0
					}
					fetchConfigs(ctx, false)
				} else {
					// 120 attempts * 15 seconds == 1800 seconds == 30 minutes
					if connectAttempt > 120 {
						// if we have been stuck trying to connect for a "long time",
						// discard the virtual connection handle and let hubConnect()
						// create a new one on the next loop
						connectAttempt = 0
						gs.hubConn = nil
					} else {
						connectAttempt += 1
						logging.Error(
							fmt.Errorf("unable to connect to stanza hub"),
							"uri", gs.clientOpt.StanzaHub,
							"attempt", connectAttempt,
						)
						gs.hubConn.Connect()
					}
				}
			} else {
				otelShutdown, sentinelShutdown = hubConnect(ctx)
			}
		}
	}
}

func fetchConfigs(ctx context.Context, skipPoll bool) {
	GetServiceConfig(ctx, skipPoll)
	GetDecoratorConfigs(ctx, skipPoll)
	if gs.httpOutboundHandler != nil {
		gsLock.Lock()
		gs.httpOutboundHandler.SetCustomerId(gs.svcConfig.GetCustomerId())
		gs.httpOutboundHandler.SetQuotaServiceClient(gs.hubQuotaClient)
		for d := range gs.decoratorConfig {
			gs.httpOutboundHandler.SetDecoratorConfig(d, gs.decoratorConfig[d])
		}
		gsLock.Unlock()
	}
	if gs.httpInboundHandler != nil {
		gsLock.Lock()
		gs.httpInboundHandler.SetCustomerId(gs.svcConfig.GetCustomerId())
		gs.httpInboundHandler.SetQuotaServiceClient(gs.hubQuotaClient)
		for d := range gs.decoratorConfig {
			gs.httpInboundHandler.SetDecoratorConfig(d, gs.decoratorConfig[d])
		}
		gsLock.Unlock()
	}
}
