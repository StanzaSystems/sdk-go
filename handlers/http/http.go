package http

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
	"google.golang.org/grpc/metadata"
)

const (
	instrumentationName    = "github.com/StanzaSystems/sdk-go/handlers/http"
	instrumentationVersion = "0.0.1" // TODO: stanza sdk-go version/build number to go here

	MAX_QUOTA_WAIT               = 1 * time.Second
	BATCH_TOKEN_CONSUME_INTERVAL = 200 * time.Millisecond
)

var (
	cachedLeases        = make(map[string][]*hubv1.TokenLease)
	cachedLeasesLock    = make(map[string]*sync.RWMutex)
	cachedLeasesGranted = make(map[string]time.Time)
	cachedLeasesRenew   = make(map[string]bool)

	consumedLeases     = []string{}
	consumedLeasesLock = &sync.RWMutex{}
	consumedLeasesInit sync.Once
)

// TODO: Implement a background poller for renewing cached leases per SDK spec:
//   When most cached tokens have been used (80% of allocated tokens) OR most (80% or more) of the cached tokens are
//   within 2 seconds of expiry, then another call to GetTokenLease must be performed and additional tokens stored
//   locally for use. Background calls such as these should specify only the Decorator name and client ID, and omit
//   the remaining parameters to the GetTokenLease endpoint (the Stanza Hub will return a set of tokens that matches
//   the statistical distribution of tokens used by this client).
// This poller should also remove expired tokens.

func checkQuota(apikey string, dc *hubv1.DecoratorConfig, qsc hubv1.QuotaServiceClient, tlr *hubv1.GetTokenLeaseRequest) (bool, string) {
	consumedLeasesInit.Do(func() {
		ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		go batchTokenConsumer(ctx, apikey, qsc)
	})

	dec := tlr.Selector.GetDecoratorName()
	if dc.GetCheckQuota() && qsc != nil {
		if _, ok := cachedLeases[dec]; !ok {
			cachedLeases[dec] = []*hubv1.TokenLease{}
			cachedLeasesLock[dec] = &sync.RWMutex{}
			cachedLeasesGranted[dec] = time.Time{}
			cachedLeasesRenew[dec] = false
		}

		if len(tlr.GetSelector().GetTags()) == 0 { // fully skip using cached leases if Quota Tags are specified
			cachedLeasesLock[dec].Lock()
			if len(cachedLeases[dec]) > 0 {
				// do we have a cachedLease at the right Feature+PriorityBoost?
				//      (priority_boost is less than or equal to the priority_boost of the current request)
				//   check expiration: time.Now().Add(time.Duration(cachedLeases[dec][x].GetDurationMsec()) * time.Millisecond)
				//      consume token
				//          (The SetTokenLeaseConsumed endpoint will accept multiple tokens for batching notifications.)
				//      remove from token cache
				//      cachedLeasesLock[dec].Unlock()
				//      return true
				logging.Debug("TODO: handle consuming of cached leases")

				// don't worry about removing expired tokens, the background poller should handle this
			}
			// No cached lease available for Feature+PriorityBoost; unlock and proceed to make a GetTokenLease request below
			cachedLeasesLock[dec].Unlock()
		}

		md := metadata.New(map[string]string{"x-stanza-key": apikey})
		ctx, cancel := context.WithTimeout(context.Background(), MAX_QUOTA_WAIT)
		defer cancel()

		resp, err := qsc.GetTokenLease(metadata.NewOutgoingContext(ctx, md), tlr)
		if err != nil {
			logging.Error(err)
			// TODO: Implement Error Handling as specified in SDK spec:
			// If quota is required and the Stanza hub is unresponsive or does not return a valid response,
			// then the SDK should do the following:
			// - time out after 300 milliseconds (and record as a failure in metrics exported to Stanza)
			//   This should be logged as a WARNING.
			// - if more than 10% of quota requests time out in a one-second period, then the SDK should fail open and
			//   stop waiting for quota from Stanza.
			//   This should be logged as an ERROR.
			// - back off for one second, and then attempt to fetch quota for 1% of requests. If over 90% of those
			//   requests are successful, ramp up to 5%, 10%, 25%, 50% and 100% over successive seconds.
			//   Re-enablement should be logged at INFO.
			return true, "" // just fail open (for now)
		}
		leases := resp.GetLeases()
		if leases == nil {
			logging.Error(fmt.Errorf("stanza-hub returned nil leases, failing open"))
			return true, "" // unexpected error! Leases should never be nil, fail open and return true (for now)
		}
		if len(leases) == 0 {
			return false, "" // not an error, there were no leases available
		}
		if len(leases[1:]) > 0 {
			logging.Debug("obtained new batch of cacheable leases",
				"decorator", dec,
				"count", len(leases[1:]))
			// lock
			// cache extra leases
			//   cachedLeases[dec] = append(cachedLeases[dec], leases[1:])
			//   cachedLeasesExpire[dec] = time.Now()
			// unlock
			// OR just throw all these leases into a queue to be added to the cache???
		}

		// TODO:
		// Should I check for leases[0].GetFeature() != tlr.Selector.GetFeatureName()?
		// If yes, what to do in case of this error?

		// Consume first token from leases (not cached, doesn't require locking)
		go consumeLease(dec, leases[0])
		return true, leases[0].Token
	}
	return true, ""
}

func consumeLease(dec string, lease *hubv1.TokenLease) {
	consumedLeasesLock.Lock()
	consumedLeases = append(consumedLeases, lease.GetToken())
	consumedLeasesLock.Unlock()
	logging.Debug("consumed quota lease",
		"decorator", dec,
		"feature", lease.GetFeature(),
		"priority", lease.GetPriorityBoost())
}

func batchTokenConsumer(ctx context.Context, apikey string, qsc hubv1.QuotaServiceClient) {
	md := metadata.New(map[string]string{"x-stanza-key": apikey})
	for {
		select {
		case <-ctx.Done():
			if qsc != nil {
				// (attempt to) flush consumed token leases to hub when we exit
				ctx, cancel := context.WithTimeout(context.Background(), MAX_QUOTA_WAIT)
				defer cancel()
				qsc.SetTokenLeaseConsumed(metadata.NewOutgoingContext(ctx, md),
					&hubv1.SetTokenLeaseConsumedRequest{Tokens: consumedLeases})
			}
			return
		case <-time.After(BATCH_TOKEN_CONSUME_INTERVAL):
			if qsc != nil {
				consumedLeasesLock.Lock()
				if len(consumedLeases) == 0 {
					consumedLeasesLock.Unlock()
				} else {
					consumeTokenReq := &hubv1.SetTokenLeaseConsumedRequest{Tokens: consumedLeases}
					consumedLeases = []string{}
					consumedLeasesLock.Unlock()

					ctx, cancel := context.WithTimeout(context.Background(), MAX_QUOTA_WAIT)
					defer cancel()
					_, err := qsc.SetTokenLeaseConsumed(metadata.NewOutgoingContext(ctx, md), consumeTokenReq)
					if err != nil {
						// if our request failed, put leases back (so they will be attempted again later)
						consumedLeasesLock.Lock()
						consumedLeases = append(consumedLeases, consumeTokenReq.Tokens...)
						consumedLeasesLock.Unlock()
						logging.Error(err)
						// TODO: add an exponential backoff sleep here?
					}
				}
			}
		}
	}
}
