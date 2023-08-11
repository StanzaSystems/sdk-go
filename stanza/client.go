package stanza

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/StanzaSystems/sdk-go/otel"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
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

// Set to how often we poll Hub for a new Decorator Config
const DECORATOR_CONFIG_REFRESH_INTERVAL = 30 * time.Second
const DECORATOR_CONFIG_REFRESH_JITTER = 6 // seconds

// Must be created outside the *Startup functions (so we don't wipe these out every 5 seconds)
var (
	otelDone     = func() {}
	sentinelDone = func() {}
)

func OtelStartup(ctx context.Context) func() {
	if OtelEnabled() {
		gsLock.Lock()
		defer gsLock.Unlock()

		GetNewBearerToken(ctx)
		if !gs.otelInit {
			otelDone, _ = otel.Init(ctx,
				gs.clientOpt.Name,
				gs.clientOpt.Release,
				gs.clientOpt.Environment)

			gs.otelInit = true
			logging.Debug("initialized opentelemetry exporter")
		}
	}
	return otelDone
}

func GetNewBearerToken(ctx context.Context) bool {
	if time.Now().After(gs.bearerTokenTime.Add(jitter(BEARER_TOKEN_REFRESH_INTERVAL, BEARER_TOKEN_REFRESH_JITTER))) {
		md := metadata.New(map[string]string{"x-stanza-key": gs.clientOpt.APIKey})
		res, err := gs.hubAuthClient.GetBearerToken(
			metadata.NewOutgoingContext(ctx, md),
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
		md := metadata.New(map[string]string{"x-stanza-key": gs.clientOpt.APIKey})
		res, err := gs.hubConfigClient.GetServiceConfig(
			metadata.NewOutgoingContext(ctx, md),
			&hubv1.GetServiceConfigRequest{
				VersionSeen: gs.svcConfigVersion,
				Service: &hubv1.ServiceSelector{
					Environment: gs.clientOpt.Environment,
					Name:        gs.clientOpt.Name,
					Release:     &gs.clientOpt.Release,
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
					if err := otel.InitMetricProvider(ctx, res.GetConfig().GetMetricConfig(), gs.bearerToken); err != nil {
						errCount += 1
						logging.Error(err)
						otel.InitMetricProvider(ctx, gs.svcConfig.GetMetricConfig(), gs.bearerToken)
					} else {
						logging.Debug("accepted opentelemetry metric config",
							"version", res.GetVersion())
					}
				}
				if gs.svcConfig.GetTraceConfig().String() != res.GetConfig().GetTraceConfig().String() {
					if err := otel.InitTraceProvider(ctx, res.GetConfig().GetTraceConfig(), gs.bearerToken); err != nil {
						errCount += 1
						logging.Error(err)
						otel.InitTraceProvider(ctx, gs.svcConfig.GetTraceConfig(), gs.bearerToken)
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

func GetDecoratorConfigs(ctx context.Context, skipPoll bool) {
	if len(gs.decoratorConfig) > 0 {
		for decorator := range gs.decoratorConfig {
			if skipPoll || time.Now().After(gs.decoratorConfigTime[decorator].Add(jitter(DECORATOR_CONFIG_REFRESH_INTERVAL, DECORATOR_CONFIG_REFRESH_JITTER))) {
				GetDecoratorConfig(ctx, decorator)
			}
		}
	}
}

func GetDecoratorConfig(ctx context.Context, decorator string) {
	if _, ok := gs.decoratorConfig[decorator]; !ok {
		gs.decoratorConfig[decorator] = &hubv1.DecoratorConfig{}
		gs.decoratorConfigTime[decorator] = time.Time{}
		gs.decoratorConfigVersion[decorator] = ""
	}
	if gs.hubConfigClient == nil {
		return
	}
	md := metadata.New(map[string]string{"x-stanza-key": gs.clientOpt.APIKey})
	res, err := gs.hubConfigClient.GetDecoratorConfig(
		metadata.NewOutgoingContext(ctx, md),
		&hubv1.GetDecoratorConfigRequest{
			VersionSeen: proto.String(gs.decoratorConfigVersion[decorator]),
			Selector: &hubv1.DecoratorServiceSelector{
				Environment:    gs.clientOpt.Environment,
				DecoratorName:  decorator,
				ServiceName:    gs.clientOpt.Name,
				ServiceRelease: gs.clientOpt.Release,
			},
		},
	)
	if err != nil {
		logging.Error(err)
	}
	if res.GetConfigDataSent() {
		gsLock.Lock()
		defer gsLock.Unlock()
		gs.decoratorConfig[decorator] = res.GetConfig()
		gs.decoratorConfigTime[decorator] = time.Now()
		gs.decoratorConfigVersion[decorator] = res.GetVersion()
		logging.Debug("accepted decorator config", "decorator", decorator, "version", res.GetVersion())
	}
}

func SentinelStartup(ctx context.Context) func() {
	if SentinelEnabled() && !gs.sentinelInit {
		done, err := sentinel.Init(gs.clientOpt.Name, gs.sentinelRules)
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
