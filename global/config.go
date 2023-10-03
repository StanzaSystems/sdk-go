package global

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/StanzaSystems/sdk-go/otel"
	"github.com/StanzaSystems/sdk-go/sentinel"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// Set to longer than it takes for all active Handlers to get new OTEL Providers
const OTEL_SHUTDOWN_GRACE_PERIOD = 4 * time.Minute

// Set to less than the maximum duration of the Auth0 Bearer Token
const BEARER_TOKEN_REFRESH_INTERVAL = 4 * time.Minute
const BEARER_TOKEN_REFRESH_JITTER = 6 // seconds

// Set to how often we poll Hub for a new Service Config
const SERVICE_CONFIG_REFRESH_INTERVAL = 30 * time.Second
const SERVICE_CONFIG_REFRESH_JITTER = 6 // seconds

// Set to how often we poll Hub for a new Guard Config
const GUARD_CONFIG_REFRESH_INTERVAL = 30 * time.Second
const GUARD_CONFIG_REFRESH_JITTER = 6 // seconds

func GetServiceConfig(ctx context.Context, skipPoll bool) {
	if skipPoll || time.Now().After(gs.svcConfigTime.Add(jitter(SERVICE_CONFIG_REFRESH_INTERVAL, SERVICE_CONFIG_REFRESH_JITTER))) {
		res, err := gs.hubConfigClient.GetServiceConfig(
			metadata.NewOutgoingContext(ctx, XStanzaKey()),
			&hubv1.GetServiceConfigRequest{
				ClientId:    proto.String(GetClientID()),
				VersionSeen: gs.svcConfigVersion,
				Service: &hubv1.ServiceSelector{
					Environment: gs.svcEnvironment,
					Name:        gs.svcName,
					Release:     &gs.svcRelease,
				},
			},
		)
		if err != nil {
			logging.Error(err)
		}
		if res.GetConfigDataSent() {
			gsLock.Lock()
			defer gsLock.Unlock()
			errCount := 0
			if gs.otelInit {
				if gs.svcConfig.GetMetricConfig().String() != res.GetConfig().GetMetricConfig().String() ||
					gs.svcConfig.GetTraceConfig().String() != res.GetConfig().GetTraceConfig().String() {
					gs.svcConfig.MetricConfig = res.GetConfig().MetricConfig
					gs.svcConfig.TraceConfig = res.GetConfig().TraceConfig
					OtelStartup(ctx, true)
					logging.Debug("accepted opentelemetry config", "version", res.GetVersion())
				}
			}
			if gs.sentinelInit {
				if sc := res.GetConfig().GetSentinelConfig(); sc != nil {
					if rules := sc.GetCircuitbreakerRulesJson(); rules != "" {
						if err := os.WriteFile(gs.sentinelRules["circuitbreaker"], []byte(rules), filePerms); err != nil {
							logging.Error(err, "version", res.GetVersion())
						}
					}
					if rules := sc.GetFlowRulesJson(); rules != "" {
						if err := os.WriteFile(gs.sentinelRules["flow"], []byte(rules), filePerms); err != nil {
							logging.Error(err, "version", res.GetVersion())
						}
					}
					if rules := sc.GetIsolationRulesJson(); rules != "" {
						if err := os.WriteFile(
							gs.sentinelRules["isolation"], []byte(rules), filePerms); err != nil {
							logging.Error(err, "version", res.GetVersion())
						}
					}
					if rules := sc.GetSystemRulesJson(); rules != "" {
						if err := os.WriteFile(gs.sentinelRules["system"], []byte(rules), filePerms); err != nil {
							logging.Error(err, "version", res.GetVersion())
						}
					}
					logging.Debug("accepted sentinel config", "version", res.GetVersion())
				}
			}
			if errCount > 0 {
				logging.Error(fmt.Errorf("rejected service config"), "version", res.GetVersion())
			} else {
				gs.svcConfig = res.GetConfig()
				gs.svcConfigTime = time.Now()
				gs.svcConfigVersion = res.GetVersion()
				logging.Debug("accepted service config", "version", res.GetVersion())
			}
		}
	}
}

func GetGuardConfigs(ctx context.Context, skipPoll bool) {
	if len(gs.guardConfig) > 0 {
		for guard := range gs.guardConfig {
			if skipPoll || time.Now().After(
				gs.guardConfigTime[guard].Add(
					jitter(GUARD_CONFIG_REFRESH_INTERVAL, GUARD_CONFIG_REFRESH_JITTER))) {
				fetchGuardConfig(ctx, guard)
			}
		}
	}
}

func fetchGuardConfig(ctx context.Context, guard string) *hubv1.GuardConfig {
	gs.guardConfigLock.RLock()
	_, ok := gs.guardConfig[guard]
	gs.guardConfigLock.RUnlock()
	if !ok {
		gs.guardConfigLock.Lock()
		gs.guardConfig[guard] = nil
		gs.guardConfigTime[guard] = time.Time{}
		gs.guardConfigVersion[guard] = ""
		gs.guardConfigLock.Unlock()
	}

	if gs.hubConfigClient == nil {
		return nil
	}
	res, err := gs.hubConfigClient.GetGuardConfig(
		metadata.NewOutgoingContext(ctx, XStanzaKey()),
		&hubv1.GetGuardConfigRequest{
			VersionSeen: proto.String(gs.guardConfigVersion[guard]),
			Selector: &hubv1.GuardServiceSelector{
				Environment:    gs.svcEnvironment,
				GuardName:      guard,
				ServiceName:    gs.svcName,
				ServiceRelease: gs.svcRelease,
			},
		},
	)
	if err != nil {
		logging.Error(err)
	}
	if res.GetConfigDataSent() {
		gs.guardConfigLock.Lock()
		gs.guardConfig[guard] = res.GetConfig()
		gs.guardConfigTime[guard] = time.Now()
		gs.guardConfigVersion[guard] = res.GetVersion()
		gs.guardConfigLock.Unlock()
		logging.Debug("accepted guard config", "guard", guard, "version", res.GetVersion())
		return res.GetConfig()
	}
	return nil
}

func OtelStartup(ctx context.Context, skipPoll bool) {
	if OtelEnabled() {
		if skipPoll || time.Now().After(gs.otelTokenTime.Add(jitter(BEARER_TOKEN_REFRESH_INTERVAL, BEARER_TOKEN_REFRESH_JITTER))) {
			if gs.svcConfig.MetricConfig != nil && gs.svcConfig.TraceConfig != nil {
				res, err := gs.hubAuthClient.GetBearerToken(
					metadata.NewOutgoingContext(ctx, XStanzaKey()),
					&hubv1.GetBearerTokenRequest{})
				if err != nil {
					logging.Error(err)
					return
				}
				if res.GetBearerToken() == "" {
					logging.Error(fmt.Errorf("failed to obtain bearer token"))
					return
				}

				sc := otel.SetupConfig{
					ServiceName:        gs.svcName,
					ServiceVersion:     gs.svcRelease,
					ServiceEnvironment: gs.svcEnvironment,
					MetricConfig:       gs.svcConfig.MetricConfig,
					TraceConfig:        gs.svcConfig.TraceConfig,
					Headers: map[string]string{
						"Authorization": "Bearer " + res.GetBearerToken(),
						"User-Agent":    UserAgent(),
					},
				}
				otelShutdown, err := otel.Setup(ctx, sc)
				if err != nil {
					logging.Error(err)
					return
				}

				// Copy the (now) old "otelShutdown" function into a goroutine
				// which will run it after a grace period; we do this so we have
				// time to update any active handlers which might be using the old
				// OTEL Providers to use our new OTEL Providers
				go runAfter(ctx, gs.otelShutdown, OTEL_SHUTDOWN_GRACE_PERIOD)

				// Finalize our success
				gs.otelInit = true
				gs.otelShutdown = otelShutdown
				gs.otelTokenTime = time.Now()
			}
		}
	}
}

func SentinelStartup(ctx context.Context) {
	if SentinelEnabled() && !gs.sentinelInit {
		done, err := sentinel.Init(gs.svcName, gs.sentinelRules)
		if err != nil {
			logging.Error(err)
			return
		}
		sentinelDone := func(ctx context.Context) error {
			done()
			os.RemoveAll(gs.sentinelDatasource)
			return err
		}
		gsLock.Lock()
		gs.sentinelInit = true
		gs.sentinelShutdown = sentinelDone
		gsLock.Unlock()
		logging.Debug("initialized sentinel rules watcher")
	}
}

// Helper function to add jitter (random number of seconds) to time.Duration
func jitter(d time.Duration, i int) time.Duration {
	return d + (time.Duration(rand.Intn(i)) * time.Second)
}

// Helper function which sleeps before given function
func runAfter(ctx context.Context, funcToRun func(context.Context) error, d time.Duration) {
	time.Sleep(d)
	err := funcToRun(ctx)
	if err != nil {
		logging.Error(err)
	}
}
