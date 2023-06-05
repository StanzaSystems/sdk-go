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

var (
	otelDone     func()
	sentinelDone func()
)

// Set to less than the maximum duration of the Auth0 Bearer Token
const BEARER_TOKEN_REFRESH_INTERVAL = 4 * time.Hour
const BEARER_TOKEN_REFRESH_JITTER = 600

// Set to how often we poll Hub for a new Service Config
const SERVICE_CONFIG_REFRESH_INTERVAL = 5 * time.Minute
const SERVICE_CONFIG_REFRESH_JITTER = 60

func jitter(d time.Duration, i int) time.Duration {
	return d + (time.Duration(rand.Intn(i)) * time.Second)
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
			gsLock.Lock()
			defer gsLock.Unlock()
			gs.bearerToken = res.GetBearerToken()
			gs.bearerTokenTime = time.Now()
			logging.Debug("successfully obtained bearer token")
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
			// TODO: compare versus current configure and trigger OTEL/Sentinel reloads if needed
			gsLock.Lock()
			defer gsLock.Unlock()
			gs.svcConfig = res.GetConfig()
			gs.svcConfigTime = time.Now()
			gs.svcConfigVersion = res.GetVersion()
			logging.Debug("successfully retrieved service config", "version", res.GetVersion())
		}
	}
}

func OtelStartup(ctx context.Context) func() {
	if OtelEnabled() {
		if !gs.otelConnected || GetNewBearerToken(ctx) {
			// TODO: connect trace and metric exporters separately
			if gs.svcConfig != nil &&
				gs.svcConfig.GetTraceConfig() != nil &&
				gs.svcConfig.GetMetricConfig() != nil {
				otelDone, _ = otel.Init(ctx,
					gs.clientOpt.Name,
					gs.clientOpt.Release,
					gs.clientOpt.Environment,
					gs.bearerToken)

				gsLock.Lock()
				defer gsLock.Unlock()
				gs.otelConnected = true
				gs.otelConnectedTime = time.Now()
				logging.Debug("successfully connected opentelemetry exporter")
			}
		}
	}
	return otelDone
}

func SentinelStartup(ctx context.Context) func() {
	if SentinelEnabled() {
		if !gs.sentinelConnected { // or X amount of time has passed
			if gs.svcConfig != nil {
				sc := gs.svcConfig.GetSentinelConfig()
				if sc != nil {
					// TODO: should setup datasource per type
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
	}
	return func() {
		sentinelDone()
		os.RemoveAll(gs.sentinelDatasource)
	}
}
