package global

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const (
	MIN_POLLING_TIME = 15 * time.Second

	filePerms = 0660
)

type state struct {
	clientId        uuid.UUID
	svcKey          string
	svcName         string
	svcEnvironment  string
	svcRelease      string
	hubURI          string
	hubConn         *grpc.ClientConn
	hubAuthClient   hubv1.AuthServiceClient
	hubConfigClient hubv1.ConfigServiceClient
	hubQuotaClient  hubv1.QuotaServiceClient

	// HTTP
	// httpInboundHandler  *httphandler.InboundHandler
	// httpOutboundHandler *httphandler.OutboundHandler

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

func NewState(ctx context.Context, hubUri, svcKey, svcName, svcEnv, svcRel string, decorators []string) func() {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	initOnce.Do(func() {
		// prepare for global state mutation
		gsLock.Lock()

		// initialize new global state
		gs = state{
			hubURI:                      hubUri,
			svcKey:                      svcKey,
			svcName:                     svcName,
			svcEnvironment:              svcEnv,
			svcRelease:                  svcRel,
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

		if len(decorators) > 0 {
			for _, decorator := range decorators {
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
			logging.Debug("disconnected from stanza hub", "uri", gs.hubURI)
		}
	}
}

func GetCustomerID() string {
	return gs.svcConfig.GetCustomerId()
}

func GetClientID() string {
	return gs.clientId.String()
}

func GetServiceKey() string {
	return gs.svcKey
}

func GetServiceName() string {
	return gs.svcName
}

func GetServiceEnvironment() string {
	return gs.svcEnvironment
}

func GetServiceRelease() string {
	return gs.svcRelease
}

func QuotaServiceClient() hubv1.QuotaServiceClient {
	return gs.hubQuotaClient
}

func DecoratorConfig(decorator string) *hubv1.DecoratorConfig {
	if dc, ok := gs.decoratorConfig[decorator]; ok {
		return dc
	}
	return nil
}
