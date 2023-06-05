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

func BearerTokenRefresh(s int) time.Duration {
	return BEARER_TOKEN_REFRESH_INTERVAL + time.Duration(rand.Intn(s))*time.Second
}

func GetBearerToken(ctx context.Context) {
	if time.Now().After(gs.bearerTokenTime.Add(BearerTokenRefresh(600))) {
		md := metadata.New(map[string]string{"x-stanza-key": gs.clientOpt.APIKey})
		res, err := gs.hubAuthClient.GetBearerToken(
			metadata.NewOutgoingContext(ctx, md),
			&hubv1.GetBearerTokenRequest{})
		if err != nil {
			logging.Error(err)
		}
		if res.GetBearerToken() != "" {
			logging.Debug("successfully obtained bearer token")
			gs.bearerToken = res.GetBearerToken()
			gs.bearerTokenTime = time.Now()
		}
	}
}

func GetServiceConfig(ctx context.Context) {
	if gs.svcConfigVersion == "" { // or X amount of time has passed
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
			logging.Debug("successfully retrieved service config", "version", res.GetVersion())
			gs.svcConfig = res.GetConfig()
			gs.svcConfigTime = time.Now()
			gs.svcConfigVersion = res.GetVersion()
		}
	}
}

func OtelStartup(ctx context.Context) func() {
	if OtelEnabled() {
		if !gs.otelConnected { // or X amount of time has passed
			// TODO: connect trace and metric exporters separately
			if gs.svcConfig != nil &&
				gs.svcConfig.GetTraceConfig() != nil &&
				gs.svcConfig.GetMetricConfig() != nil {
				otelDone, _ = otel.Init(ctx,
					gs.clientOpt.Name,
					gs.clientOpt.Release,
					gs.clientOpt.Environment,
					gs.bearerToken)
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
