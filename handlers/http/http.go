package http

import (
	"context"
	"sync"
	"time"

	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
	"google.golang.org/grpc/metadata"
)

const (
	instrumentationName    = "github.com/StanzaSystems/sdk-go/handlers/http"
	instrumentationVersion = "0.0.1" // TODO: stanza sdk-go version/build number to go here

	MAX_QUOTA_WAIT = time.Duration(2) * time.Second
)

var (
	availableLeases       = make(map[*hubv1.GetTokenLeaseRequest][]*hubv1.TokenLease)
	availableLeasesLock   = make(map[*hubv1.GetTokenLeaseRequest]*sync.RWMutex)
	availableLeasesExpire = make(map[*hubv1.GetTokenLeaseRequest]time.Time)
)

func checkQuota(apikey string, dc *hubv1.DecoratorConfig, qsc hubv1.QuotaServiceClient, tlr *hubv1.GetTokenLeaseRequest) bool {
	quotaSuccess := true // fail open in the face of errors
	if dc.GetCheckQuota() && qsc != nil {
		if _, ok := availableLeases[tlr]; !ok {
			availableLeases[tlr] = []*hubv1.TokenLease{}
			availableLeasesLock[tlr] = &sync.RWMutex{}
			availableLeasesExpire[tlr] = time.Time{}
		}
		ctx, cancel := context.WithTimeout(context.Background(), MAX_QUOTA_WAIT)
		defer cancel()
		md := metadata.New(map[string]string{"x-stanza-key": apikey})

		availableLeasesLock[tlr].Lock()
		if len(availableLeases[tlr]) == 0 || time.Now().After(availableLeasesExpire[tlr]) {
			resp, err := qsc.GetTokenLease(metadata.NewOutgoingContext(ctx, md), tlr)
			if err != nil {
				logging.Error(err)
			} else {
				if resp.GetLeases() == nil || len(resp.GetLeases()) == 0 {
					quotaSuccess = false // not an error, there are no leases available
				} else {
					availableLeases[tlr] = resp.GetLeases()
					availableLeasesExpire[tlr] = time.Now().Add(time.Duration(resp.GetLeases()[0].GetDurationMsec()) * time.Millisecond)
					logging.Debug("obtained new batch of quota leases",
						"env", tlr.Selector.GetEnvironment(),
						"decorator", tlr.Selector.GetDecoratorName(),
						"feature", tlr.Selector.GetFeatureName(),
						"count", len(availableLeases[tlr]))
				}
			}
		}
		if quotaSuccess && len(availableLeases[tlr]) > 0 {
			go qsc.SetTokenLeaseConsumed(metadata.NewOutgoingContext(ctx, md), &hubv1.SetTokenLeaseConsumedRequest{Tokens: []string{availableLeases[tlr][0].Token}})
			logging.Debug("consumed quota lease",
				"env", tlr.Selector.GetEnvironment(),
				"decorator", tlr.Selector.GetDecoratorName(),
				"feature", tlr.Selector.GetFeatureName(),
				"remaining", len(availableLeases[tlr][1:]))
			availableLeases[tlr] = availableLeases[tlr][1:]
		}
		availableLeasesLock[tlr].Unlock()
	}
	return quotaSuccess
}
