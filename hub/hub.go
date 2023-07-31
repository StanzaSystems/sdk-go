package hub

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/StanzaSystems/sdk-go/logging"
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	MAX_QUOTA_WAIT               = 1 * time.Second
	CACHED_LEASE_CHECK_INTERVAL  = 200 * time.Millisecond // TODO: what should this be set to?
	BATCH_TOKEN_CONSUME_INTERVAL = 200 * time.Millisecond // TODO: what should this be set to?
)

var (
	waitingLeases     = make(map[string][]*hubv1.TokenLease)
	waitingLeasesLock = make(map[string]*sync.RWMutex)

	cachedLeases     = make(map[string][]*hubv1.TokenLease)
	cachedLeasesLock = make(map[string]*sync.RWMutex)
	cachedLeasesUsed = make(map[string]int)
	cachedLeasesReq  = make(map[string]*hubv1.GetTokenLeaseRequest)
	cachedLeasesInit sync.Once

	consumedLeases     = []string{}
	consumedLeasesLock = &sync.RWMutex{}
	consumedLeasesInit sync.Once
)

func CheckQuota(apikey string, dc *hubv1.DecoratorConfig, qsc hubv1.QuotaServiceClient, tlr *hubv1.GetTokenLeaseRequest) (bool, string) {
	// Start a background batch token consumer (the first time checkQuota is called)
	consumedLeasesInit.Do(func() { go batchTokenConsumer(apikey, qsc) })

	dec := tlr.Selector.GetDecoratorName()
	if dc.GetCheckQuota() && qsc != nil {
		if _, ok := cachedLeases[dec]; !ok {
			cachedLeases[dec] = []*hubv1.TokenLease{}
			cachedLeasesLock[dec] = &sync.RWMutex{}
			cachedLeasesUsed[dec] = 0
			cachedLeasesReq[dec] = &hubv1.GetTokenLeaseRequest{
				Selector: &hubv1.DecoratorFeatureSelector{
					Environment:   tlr.Selector.GetEnvironment(),
					DecoratorName: tlr.Selector.GetDecoratorName(),
				},
				ClientId: tlr.ClientId,
			}
		}
		if _, ok := waitingLeases[dec]; !ok {
			waitingLeases[dec] = []*hubv1.TokenLease{}
			waitingLeasesLock[dec] = &sync.RWMutex{}
		}

		if len(tlr.GetSelector().GetTags()) == 0 { // fully skip using cached leases if Quota Tags are specified
			cachedLeasesLock[dec].Lock()
			if len(cachedLeases[dec]) > 0 {
				for k, tl := range cachedLeases[dec] {
					if tl.GetFeature() == tlr.GetSelector().GetFeatureName() {
						if tl.GetPriorityBoost() <= tlr.GetPriorityBoost() {
							if time.Now().Before(tl.GetExpiresAt().AsTime()) {
								// we have a cached lease for the given feature, at the right priority, which hasn't expired
								newCache := append(cachedLeases[dec][:k], cachedLeases[dec][k+1:]...)
								cachedLeases[dec] = newCache
								cachedLeasesUsed[dec] += 1
								cachedLeasesLock[dec].Unlock()
								return true, tl.Token
							}
						}
					}
				}
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
			// If quota is required and the Stanza hub is unresponsive or does not return a valid
			// response, then the SDK should do the following:
			// - time out after 300 milliseconds (and record as a failure in metrics exported to Stanza)
			//   This should be logged as a WARNING.
			// - if more than 10% of quota requests time out in a one-second period, then the SDK should
			//   fail open and stop waiting for quota from Stanza.
			//   This should be logged as an ERROR.
			// - back off for one second, and then attempt to fetch quota for 1% of requests. If over 90%
			//   of those requests are successful, ramp up to 5%, 10%, 25%, 50% and 100% over successive
			//   seconds.
			//   Re-enablement should be logged at INFO.
			return true, "" // just fail open (for now)
		}
		leases := resp.GetLeases()
		if len(leases) == 0 {
			return false, "" // not an error, there were no leases available
		}
		if len(leases[1:]) > 0 {
			// Start a background cached lease manager (the first time we get extra leases from Stanza Hub)
			cachedLeasesInit.Do(func() { go cachedLeaseManager(apikey, qsc) })

			logging.Debug("obtained new batch of cacheable leases", "decorator", dec, "count", len(leases[1:]))
			for _, lease := range leases[1:] {
				if lease.ExpiresAt == nil {
					lease.ExpiresAt = timestamppb.New(time.Now().Add(time.Duration(lease.DurationMsec) * time.Millisecond))
				}
			}
			// use a separate "waiting leases" lock here as we don't need/want to block a request on contention for
			// the higher volume / harder to get "cached leases" lock
			waitingLeasesLock[dec].Lock()
			waitingLeases[dec] = append(waitingLeases[dec], leases[1:]...)
			waitingLeasesLock[dec].Unlock()
		}

		// Consume first token from leases (not cached, so this doesn't require the cached leases lock)
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
		"weight", lease.GetWeight(),
		"priority_boost", lease.GetPriorityBoost())
}

func batchTokenConsumer(apikey string, qsc hubv1.QuotaServiceClient) {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
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

func cachedLeaseManager(apikey string, qsc hubv1.QuotaServiceClient) {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	md := metadata.New(map[string]string{"x-stanza-key": apikey})
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(CACHED_LEASE_CHECK_INTERVAL):
			for dec := range cachedLeases {
				cachedLeasesLock[dec].Lock()
				newCache := []*hubv1.TokenLease{}
				cachedLeaseCount := len(cachedLeases[dec])
				expiringLeaseCount := 0

				// Check for and remove any expired leases
				for k, tl := range cachedLeases[dec] {
					if time.Now().Before(tl.GetExpiresAt().AsTime()) {
						newCache = append(newCache, cachedLeases[dec][k])
					} else {
						cachedLeasesUsed[dec] += 1
					}
				}

				// Check for number of leases within 2 seconds of expiring
				for _, tl := range newCache {
					if time.Now().Before(tl.GetExpiresAt().AsTime().Add(-2 * time.Second)) {
						expiringLeaseCount += 1
					}
				}

				// Add any additional leases waiting to be cached now
				waitingLeasesLock[dec].Lock()
				if len(waitingLeases[dec]) > 0 {
					newCache = append(newCache, waitingLeases[dec]...)
					cachedLeaseCount += len(waitingLeases[dec])
					cachedLeasesUsed[dec] = 0
					waitingLeases[dec] = []*hubv1.TokenLease{}
				}
				waitingLeasesLock[dec].Unlock()

				// Make a GetTokenLease request if >80% of our tokens are already used (or expiring soon)
				if qsc != nil {
					if float32((cachedLeaseCount-expiringLeaseCount)/(cachedLeaseCount+cachedLeasesUsed[dec])) < 0.2 {
						go func() {
							ctx, cancel := context.WithTimeout(context.Background(), CACHED_LEASE_CHECK_INTERVAL)
							defer cancel()
							resp, err := qsc.GetTokenLease(metadata.NewOutgoingContext(ctx, md), cachedLeasesReq[dec])
							if err != nil {
								logging.Error(err)
							}
							if len(resp.GetLeases()) > 0 {
								waitingLeasesLock[dec].Lock()
								waitingLeases[dec] = append(waitingLeases[dec], resp.GetLeases()...)
								waitingLeasesLock[dec].Unlock()
							}
						}()
					}
				}

				// Update the cached leases store
				cachedLeases[dec] = newCache
				cachedLeasesLock[dec].Unlock()
			}
		}
	}
}

func ValidateTokens(apikey, environment, decorator string, dc *hubv1.DecoratorConfig, qsc hubv1.QuotaServiceClient, tokens []string) bool {
	if !dc.GetValidateIngressTokens() {
		return true // if we weren't asked to validate ingress tokens, don't
	}
	if len(tokens) == 0 {
		logging.Warn("validate ingress tokens was specified, but no tokens were found", "decorator", decorator)
		return false // fail fast in the case where we are supposed to validate, but no tokens found
	}

	ds := &hubv1.DecoratorSelector{Environment: environment, Name: decorator}
	vtr := &hubv1.ValidateTokenRequest{Tokens: tokenInfos(tokens, ds)}

	md := metadata.New(map[string]string{"x-stanza-key": apikey})
	ctx, cancel := context.WithTimeout(context.Background(), MAX_QUOTA_WAIT)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			logging.Error(ctx.Err())
			return true // deadline reached, log error and fail open
		default:
			resp, err := qsc.ValidateToken(metadata.NewOutgoingContext(ctx, md), vtr)
			if err != nil {
				logging.Error(err)
				return true // error from Stanza Hub, log error and fail open
			}
			for _, t := range resp.GetTokensValid() {
				if !t.Valid {
					return false
				}
			}
			return true
		}
	}
}

func tokenInfos(tokens []string, ds *hubv1.DecoratorSelector) (ti []*hubv1.TokenInfo) {
	for _, t := range tokens {
		ti = append(ti, &hubv1.TokenInfo{
			Token:     t,
			Decorator: ds,
		})
	}
	return ti
}
