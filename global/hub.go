package global

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"time"

	"buf.build/gen/go/stanza/apis/grpc/go/stanza/hub/v1/hubv1grpc"
	"github.com/StanzaSystems/sdk-go/ca"
	"github.com/StanzaSystems/sdk-go/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func hubConnect(ctx context.Context) {
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
		grpc.WithUserAgent(UserAgent()),
		// todo: add keepalives, backoff config, etc
	}
	hubConn, err := grpc.Dial(gs.hubURI, opts...)
	if err != nil {
		logging.Error(err,
			"msg", "failed to connect to stanza hub",
			"url", gs.hubURI)
	} else {
		gsLock.Lock()
		gs.hubConn = hubConn
		gs.hubAuthClient = hubv1grpc.NewAuthServiceClient(hubConn)
		gs.hubConfigClient = hubv1grpc.NewConfigServiceClient(hubConn)
		gs.hubQuotaClient = hubv1grpc.NewQuotaServiceClient(hubConn)
		gsLock.Unlock()

		// attempt to establish hub connection (doesn't block)
		gs.hubConn.Connect()

		// block, waiting for up to 10 seconds for hub connection
		ctxWait, ctxWaitCancel := context.WithTimeout(ctx, 10*time.Second)
		defer ctxWaitCancel()
		gs.hubConn.WaitForStateChange(ctxWait, connectivity.Connecting)
		if gs.hubConn.GetState() == connectivity.Ready {
			logging.Info("connected to stanza hub", "uri", gs.hubURI)
			GetServiceConfig(ctx, true)
			GetGuardConfigs(ctx, true)
			OtelStartup(ctx, true)
			SentinelStartup(ctx)
		}
	}
}

func hubPoller(ctx context.Context, pollInterval time.Duration) {
	connectAttempt := 0
	for {
		select {
		case <-ctx.Done():
			gs.otelShutdown(ctx)
			gs.sentinelShutdown(ctx)
			return
		case <-time.After(pollInterval):
			if gs.hubConn != nil {
				if gs.hubConn.GetState() == connectivity.Ready {
					if connectAttempt > 0 {
						logging.Info(
							"connected to stanza hub",
							"uri", gs.hubURI,
							"attempt", connectAttempt,
						)
						connectAttempt = 0
					}
					GetServiceConfig(ctx, false)
					GetGuardConfigs(ctx, false)
					OtelStartup(ctx, false)
					SentinelStartup(ctx)
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
							"uri", gs.hubURI,
							"attempt", connectAttempt,
						)
						host, _, _ := net.SplitHostPort(gs.hubURI)
						_, err := net.LookupHost(host)
						if err != nil {
							logging.Error(err)
						}
						gs.hubConn.Connect()
					}
				}
			} else {
				hubConnect(ctx)
			}
		}
	}
}
