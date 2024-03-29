package global

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"buf.build/gen/go/stanza/apis/grpc/go/stanza/hub/v1/hubv1grpc"
	hubv1 "buf.build/gen/go/stanza/apis/protocolbuffers/go/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	instrumentationName    = "github.com/StanzaSystems/sdk-go"
	instrumentationVersion = "0.0.5-beta"

	MIN_POLLING_TIME = 15 * time.Second

	filePerms = 0660
)

type state struct {
	clientId       uuid.UUID
	svcKey         string
	svcName        string
	svcEnvironment string
	svcRelease     string
	hubURI         string

	// stored after hubConnect success
	hubConn         *grpc.ClientConn
	hubAuthClient   hubv1grpc.AuthServiceClient
	hubConfigClient hubv1grpc.ConfigServiceClient
	hubQuotaClient  hubv1grpc.QuotaServiceClient

	// stored from GetServiceConfig polling
	svcConfig        *hubv1.ServiceConfig
	svcConfigTime    time.Time
	svcConfigVersion string

	// stored from GetGuardConfig polling
	guardConfig        map[string]*hubv1.GuardConfig
	guardConfigTime    map[string]time.Time
	guardConfigVersion map[string]string
	guardConfigLock    *sync.RWMutex

	// otel
	otelInit         bool
	otelShutdown     func(context.Context) error
	otelTokenTime    time.Time
	otelStanzaMeter  *StanzaMeter
	otelStanzaTracer *trace.Tracer

	// sentinel
	sentinelInit       bool
	sentinelShutdown   func(context.Context) error
	sentinelDatasource string
	sentinelRules      map[string]string
	sentinelRulesLock  *sync.RWMutex
}

var (
	gs       = state{}
	gsLock   = &sync.RWMutex{}
	initOnce sync.Once
)

func NewState(ctx context.Context, hubUri, svcKey, svcName, svcEnv, svcRel string, guards []string) func() {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// initialize new global state
	initOnce.Do(func() {
		gsLock.Lock()
		gs = state{
			hubURI:             hubUri,
			svcKey:             svcKey,
			svcName:            svcName,
			svcEnvironment:     svcEnv,
			svcRelease:         svcRel,
			clientId:           uuid.New(),
			hubConn:            nil,
			svcConfig:          &hubv1.ServiceConfig{},
			svcConfigTime:      time.Time{},
			svcConfigVersion:   "",
			guardConfig:        make(map[string]*hubv1.GuardConfig),
			guardConfigTime:    make(map[string]time.Time),
			guardConfigVersion: make(map[string]string),
			guardConfigLock:    &sync.RWMutex{},
			otelInit:           false,
			otelShutdown:       func(context.Context) error { return nil },
			otelTokenTime:      time.Time{},
			otelStanzaMeter:    NewStanzaMeter(),
			otelStanzaTracer:   NewStanzaTracer(),
			sentinelInit:       false,
			sentinelShutdown:   func(context.Context) error { return nil },
			sentinelRulesLock:  &sync.RWMutex{},
		}
		gsLock.Unlock()

		// pre-create empty sentinel rules files
		gs.sentinelRulesLock.Lock()
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
		gs.sentinelRulesLock.Unlock()

		if len(guards) > 0 {
			for _, guard := range guards {
				gs.guardConfigLock.Lock()
				gs.guardConfig[guard] = nil
				gs.guardConfigTime[guard] = time.Time{}
				gs.guardConfigVersion[guard] = ""
				gs.guardConfigLock.Unlock()
			}
		}

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
	gsLock.RLock()
	defer gsLock.RUnlock()
	return gs.svcConfig.GetCustomerId()
}

func GetClientID() string {
	return gs.clientId.String()
}

func GetServiceKey() string {
	return gs.svcKey
}

func XStanzaKey() metadata.MD {
	return metadata.New(map[string]string{"x-stanza-key": gs.svcKey})
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

func GetStanzaMeter() *StanzaMeter {
	return gs.otelStanzaMeter
}

func GetStanzaTracer() *trace.Tracer {
	return gs.otelStanzaTracer
}

func GetGuardConfig(ctx context.Context, guard string) (*hubv1.GuardConfig, hubv1.Config, error) {
	gs.guardConfigLock.RLock()
	gc, ok := gs.guardConfig[guard]
	gs.guardConfigLock.RUnlock()
	if ok && gc != nil {
		return gc, hubv1.Config_CONFIG_CACHED_OK, nil
	}
	return fetchGuardConfig(ctx, guard)
}

func QuotaServiceClient() hubv1grpc.QuotaServiceClient {
	gsLock.RLock()
	defer gsLock.RUnlock()
	return gs.hubQuotaClient
}

func InstrumentationName() string {
	return instrumentationName
}

func InstrumentationMetricVersion() metric.MeterOption {
	return metric.WithInstrumentationVersion(instrumentationVersion)
}

func InstrumentationTraceVersion() trace.TracerOption {
	return trace.WithInstrumentationVersion(instrumentationVersion)
}

func UserAgent() string {
	return fmt.Sprintf("%s/%s StanzaGoSDK/v%s", gs.svcName, gs.svcRelease, instrumentationVersion)
}
