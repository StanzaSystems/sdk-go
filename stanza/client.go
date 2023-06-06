package stanza

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/StanzaSystems/sdk-go/otel"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/sentinel"

	"google.golang.org/grpc/metadata"
)

// Set to less than the maximum duration of the Auth0 Bearer Token
const BEARER_TOKEN_REFRESH_INTERVAL = 4 * time.Hour
const BEARER_TOKEN_REFRESH_JITTER = 600

// Set to how often we poll Hub for a new Service Config
const SERVICE_CONFIG_REFRESH_INTERVAL = 5 * time.Minute
const SERVICE_CONFIG_REFRESH_JITTER = 60

// Must be created outside the *Startup functions (so we don't wipe these out every 5 seconds)
var (
	otelDone     = func() {}
	sentinelDone = func() {}
)

func OtelStartup(ctx context.Context) func() {
	if OtelEnabled() && gs.svcConfig != nil {
		gsLock.Lock()
		defer gsLock.Unlock()
		if !gs.otelInit {
			otelDone, _ = otel.Init(ctx,
				gs.clientOpt.Name,
				gs.clientOpt.Release,
				gs.clientOpt.Environment)

			gs.otelInit = true
			logging.Debug("initialized opentelemetry exporter")
		}
		newToken := GetNewBearerToken(ctx)
		if !gs.otelMetricProviderConnected || newToken {
			if gs.svcConfig.GetMetricConfig() != nil {
				if err := otel.InitMetricProvider(ctx, gs.svcConfig.GetMetricConfig(), gs.bearerToken); err != nil {
					logging.Error(err)
				} else {
					gs.otelMetricProviderConnected = true
					if os.Getenv("STANZA_DEBUG") != "" || os.Getenv("STANZA_OTEL_DEBUG") != "" {
						logging.Debug("connected to stdout metric exporter")
					} else {
						logging.Debug("connected to opentelemetry metric collector", "url", gs.svcConfig.GetMetricConfig().GetCollectorUrl())
					}
				}
			}
		}
		if !gs.otelTraceProviderConnected || newToken {
			if gs.svcConfig.GetTraceConfig() != nil {
				if err := otel.InitTraceProvider(ctx, gs.svcConfig.GetTraceConfig(), gs.bearerToken); err != nil {
					logging.Error(err)
				} else {
					gs.otelTraceProviderConnected = true
					if os.Getenv("STANZA_DEBUG") != "" || os.Getenv("STANZA_OTEL_DEBUG") != "" {
						logging.Debug("connected to stdout trace exporter")
					} else {
						logging.Debug("connected to opentelemetry trace collector", "url", gs.svcConfig.GetTraceConfig().GetCollectorUrl())
					}
				}
			}
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

func GetServiceConfig(ctx context.Context) {
	if time.Now().After(gs.svcConfigTime.Add(jitter(SERVICE_CONFIG_REFRESH_INTERVAL, SERVICE_CONFIG_REFRESH_JITTER))) {
		md := metadata.New(map[string]string{"x-stanza-key": gs.clientOpt.APIKey})
		res, _ := gs.hubConfigClient.GetServiceConfig(
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
		if res.GetConfigDataSent() {
			gsLock.Lock()
			defer gsLock.Unlock()
			// TODO: compare versus current configure and trigger OTEL/Sentinel reloads if needed
			gs.svcConfig = res.GetConfig()
			gs.svcConfigTime = time.Now()
			gs.svcConfigVersion = res.GetVersion()
			logging.Debug("retrieved service config", "version", res.GetVersion())
		}
	}
}

func SentinelStartup(ctx context.Context) func() {
	if SentinelEnabled() && gs.svcConfig != nil {
		if !gs.sentinelConnected { // or X amount of time has passed
			sc := gs.svcConfig.GetSentinelConfig()
			if sc != nil {
				// TODO: should init datasource per type
				if sc.GetCircuitbreakerRulesJson() != "" ||
					sc.GetFlowRulesJson() != "" ||
					sc.GetIsolationRulesJson() != "" ||
					sc.GetSystemRulesJson() != "" {
					// TODO: need to add a gs.svcConfig.SentinelConfig -> gs.sentinelDatasource writer
					sentinelDone = sentinel.Init(gs.clientOpt.Name, gs.sentinelDatasource)

					gsLock.Lock()
					defer gsLock.Unlock()
					gs.sentinelConnected = true
					gs.sentinelConnectedTime = time.Now()
					logging.Debug("successfully connected sentinel watcher")
				}
			}
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
