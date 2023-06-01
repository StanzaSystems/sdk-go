package stanza

import (
	"context"
	"fmt"
	"time"

	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"google.golang.org/grpc/metadata"
)

func GetBearerToken(ctx context.Context) {
	if gs.bearerToken == "" { // or X amount of time has passed
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
		}
	}
	fmt.Println(gs.bearerToken)
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
			gs.svcConfig = *res.GetConfig()
			gs.svcConfigTime = time.Now()
			gs.svcConfigVersion = res.GetVersion()
		}
	}
	fmt.Println(gs.svcConfigVersion)
}
