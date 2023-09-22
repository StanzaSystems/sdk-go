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

// Set to less than the maximum duration of the Auth0 Bearer Token
const BEARER_TOKEN_REFRESH_INTERVAL = 4 * time.Hour
const BEARER_TOKEN_REFRESH_JITTER = 600 // seconds

// Set to how often we poll Hub for a new Service Config
const SERVICE_CONFIG_REFRESH_INTERVAL = 30 * time.Second
const SERVICE_CONFIG_REFRESH_JITTER = 6 // seconds

// Set to how often we poll Hub for a new Guard Config
const GUARD_CONFIG_REFRESH_INTERVAL = 30 * time.Second
const GUARD_CONFIG_REFRESH_JITTER = 6 // seconds

// Must be created outside the *Startup functions (so we don't wipe these out every 5 seconds)
var (
	otelDone     = func() {}
	sentinelDone = func() {}
)

func OtelStartup(ctx context.Context) func() {
	if OtelEnabled() {
		gsLock.Lock()
		defer gsLock.Unlock()

		if !gs.otelInit {
			otelDone, _ = otel.Init(ctx,
				gs.svcName,
				gs.svcRelease,
				gs.svcEnvironment)

			gs.otelInit = true
			logging.Debug("initialized opentelemetry exporter")
		}
		if gs.svcConfig != nil && GetNewBearerToken(ctx) {
			if err := otel.InitMetricProvider(ctx, gs.svcConfig.GetMetricConfig(), gs.bearerToken, UserAgent()); err != nil {
				logging.Error(err)
			} else {
				logging.Debug("accepted opentelemetry metric config",
					"version", gs.svcConfigVersion)
			}
			if err := otel.InitTraceProvider(ctx, gs.svcConfig.GetTraceConfig(), gs.bearerToken, UserAgent()); err != nil {
				logging.Error(err)
			} else {
				logging.Debug("accepted opentelemetry trace config",
					"version", gs.svcConfigVersion)
			}
		}
	}
	return otelDone
}

func GetNewBearerToken(ctx context.Context) bool {
	if time.Now().After(gs.bearerTokenTime.Add(jitter(BEARER_TOKEN_REFRESH_INTERVAL, BEARER_TOKEN_REFRESH_JITTER))) {
		res, err := gs.hubAuthClient.GetBearerToken(
			metadata.NewOutgoingContext(ctx, XStanzaKey()),
			&hubv1.GetBearerTokenRequest{})
		if err != nil {
			logging.Error(err)
		}
		if res.GetBearerToken() != "" {
			gs.bearerToken = res.GetBearerToken()
			gs.bearerTokenTime = time.Now()
			logging.Debug("obtained bearer token")
			return true
		}
	}
	return false
}

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
				if gs.svcConfig.GetMetricConfig().String() != res.GetConfig().GetMetricConfig().String() {
					if err := otel.InitMetricProvider(ctx, res.GetConfig().GetMetricConfig(), gs.bearerToken, UserAgent()); err != nil {
						errCount += 1
						logging.Error(err)
						otel.InitMetricProvider(ctx, gs.svcConfig.GetMetricConfig(), gs.bearerToken, UserAgent())
					} else {
						logging.Debug("accepted opentelemetry metric config",
							"version", res.GetVersion())
					}
				}
				if gs.svcConfig.GetTraceConfig().String() != res.GetConfig().GetTraceConfig().String() {
					if err := otel.InitTraceProvider(ctx, res.GetConfig().GetTraceConfig(), gs.bearerToken, UserAgent()); err != nil {
						errCount += 1
						logging.Error(err)
						otel.InitTraceProvider(ctx, gs.svcConfig.GetTraceConfig(), gs.bearerToken, UserAgent())
					} else {
						logging.Debug("accepted opentelemetry trace config",
							"version", res.GetVersion(),
							"sample_rate", res.GetConfig().GetTraceConfig().GetSampleRateDefault())
					}
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

func SentinelStartup(ctx context.Context) func() {
	if SentinelEnabled() && !gs.sentinelInit {
		done, err := sentinel.Init(gs.svcName, gs.sentinelRules)
		if err != nil {
			logging.Error(err)
		} else {
			sentinelDone = done
			gsLock.Lock()
			gs.sentinelInit = true
			gs.sentinelInitTime = time.Now()
			gsLock.Unlock()
			logging.Debug("initialized sentinel rules watcher")
		}
	}
	return func() {
		sentinelDone()
		os.RemoveAll(gs.sentinelDatasource)
	}
}

// Helper function to add jitter (random number of seconds) to time.Duration
func jitter(d time.Duration, i int) time.Duration {
	return d + (time.Duration(rand.Intn(i)) * time.Second)
}
