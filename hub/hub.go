package hub

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/logging"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	MAX_QUOTA_WAIT               = 1 * time.Second
	CACHED_LEASE_CHECK_INTERVAL  = 200 * time.Millisecond // TODO: what should this be set to?
	BATCH_TOKEN_CONSUME_INTERVAL = 200 * time.Millisecond // TODO: what should this be set to?

	// CheckQuota Response Codes
	CheckQuotaAllowed = iota
	CheckQuotaBlocked
	CheckQuotaSkipped  // special case of Allowed
	CheckQuotaFailOpen // special case of Allowed

	// ValidateTokens Response Codes
	ValidateTokensValid = iota
	ValidateTokensInvalid
	ValidateTokensSkipped  // special case of Valid
	ValidateTokensFailOpen // special case of Valid
)

var (
	waitingLeasesMapLock = &sync.RWMutex{}
	waitingLeases        = make(map[string][]*hubv1.TokenLease)
	waitingLeasesLock    = make(map[string]*sync.RWMutex)

	cachedLeasesMapLock = &sync.RWMutex{}
	cachedLeases        = make(map[string][]*hubv1.TokenLease)
	cachedLeasesLock    = make(map[string]*sync.RWMutex)
	cachedLeasesUsed    = make(map[string]int)
	cachedLeasesReq     = make(map[string]*hubv1.GetTokenLeaseRequest)
	cachedLeasesInit    sync.Once

	consumedLeases     = []string{}
	consumedLeasesLock = &sync.RWMutex{}
	consumedLeasesInit sync.Once

	failOpenCount = int64(0)
)

func NewTokenLeaseRequest(guard string) *hubv1.GetTokenLeaseRequest {
	return &hubv1.GetTokenLeaseRequest{
		ClientId: proto.String(global.GetClientID()),
		Selector: &hubv1.GuardFeatureSelector{
			GuardName:   guard,
			Environment: global.GetServiceEnvironment(),
		},
	}
}

func CheckQuota(ctx context.Context, tlr *hubv1.GetTokenLeaseRequest) (int, string) {
	if tlr == nil || tlr.Selector == nil {
		logging.Debug(
			"invalid token lease request, failing open",
			"count", atomic.AddInt64(&failOpenCount, 1),
		)
		return CheckQuotaFailOpen, ""
	}
	qsc := global.QuotaServiceClient()
	if qsc == nil {
		logging.Debug(
			"invalid quota service client, failing open",
			"count", atomic.AddInt64(&failOpenCount, 1),
		)
		return CheckQuotaFailOpen, ""
	}
	guard := tlr.GetSelector().GetGuardName()
	gc := global.GetGuardConfig(ctx, guard)
	if gc == nil {
		logging.Debug(
			"invalid guard config, failing open",
			"count", atomic.AddInt64(&failOpenCount, 1),
		)
		return CheckQuotaFailOpen, ""
	}
	if !gc.GetCheckQuota() {
		return CheckQuotaSkipped, ""
	}

	// start a background batch token consumer
	consumedLeasesInit.Do(func() { go batchTokenConsumer() })

	cachedLeasesMapLock.RLock()
	_, cachedLeasesExists := cachedLeases[guard]
	cachedLeasesMapLock.RUnlock()
	if !cachedLeasesExists {
		cachedLeasesMapLock.Lock()
		cachedLeases[guard] = []*hubv1.TokenLease{}
		cachedLeasesLock[guard] = &sync.RWMutex{}
		cachedLeasesUsed[guard] = 0
		cachedLeasesReq[guard] = &hubv1.GetTokenLeaseRequest{
			Selector: &hubv1.GuardFeatureSelector{
				Environment: tlr.GetSelector().GetEnvironment(),
				GuardName:   tlr.GetSelector().GetGuardName(),
			},
			ClientId: tlr.ClientId,
		}
		cachedLeasesMapLock.Unlock()
	}
	waitingLeasesMapLock.RLock()
	_, waitingLeasesExists := waitingLeases[guard]
	waitingLeasesMapLock.RUnlock()
	if !waitingLeasesExists {
		waitingLeasesMapLock.Lock()
		waitingLeases[guard] = []*hubv1.TokenLease{}
		waitingLeasesLock[guard] = &sync.RWMutex{}
		waitingLeasesMapLock.Unlock()

	}

	if len(tlr.GetSelector().GetTags()) == 0 { // fully skip using cached leases if Quota Tags are specified
		cachedLeasesLock[guard].RLock()
		cachedLeaseLen := len(cachedLeases[guard])
		cachedLeasesLock[guard].RUnlock()
		if cachedLeaseLen > 0 {
			cachedLeasesLock[guard].Lock()
			for k, tl := range cachedLeases[guard] {
				if tl.GetFeature() == tlr.GetSelector().GetFeatureName() {
					if tl.GetPriorityBoost() <= tlr.GetPriorityBoost() {
						if time.Now().Before(tl.GetExpiresAt().AsTime()) {
							// We have a cached lease for the given feature, at the right priority,
							// which hasn't expired; remove from cache, unlock, and return cached token
							newCache := append(cachedLeases[guard][:k], cachedLeases[guard][k+1:]...)
							cachedLeases[guard] = newCache
							cachedLeasesUsed[guard] += 1
							cachedLeasesLock[guard].Unlock()
							return CheckQuotaAllowed, tl.Token
						}
					}
				}
			}
			// No cached lease available for Feature+PriorityBoost;
			// unlock and proceed to make a GetTokenLease request below
			cachedLeasesLock[guard].Unlock()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), MAX_QUOTA_WAIT)
	defer cancel()

	resp, err := qsc.GetTokenLease(metadata.NewOutgoingContext(ctx, global.XStanzaKey()), tlr)
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
		return CheckQuotaFailOpen, "" // just fail open (for now)
	}
	leases := resp.GetLeases()
	if len(leases) == 0 {
		return CheckQuotaBlocked, "" // not an error, there were no leases available
	}
	if len(leases[1:]) > 0 {
		// Start a background cached lease manager (the first time we get extra leases from Stanza Hub)
		cachedLeasesInit.Do(func() { go cachedLeaseManager() })

		logging.Debug("obtained new batch of cacheable leases", "guard", guard, "count", len(leases[1:]))
		for _, lease := range leases[1:] {
			if lease.ExpiresAt == nil {
				lease.ExpiresAt = timestamppb.New(time.Now().Add(time.Duration(lease.DurationMsec) * time.Millisecond))
			}
		}
		// use a separate "waiting leases" lock here as we don't need/want to block a request on contention for
		// the higher volume / harder to get "cached leases" lock
		waitingLeasesLock[guard].Lock()
		waitingLeases[guard] = append(waitingLeases[guard], leases[1:]...)
		waitingLeasesLock[guard].Unlock()
	}

	// Consume first token from leases (not cached, so this doesn't require the cached leases lock)
	go consumeLease(guard, leases[0])
	return CheckQuotaAllowed, leases[0].Token
}

func consumeLease(guard string, lease *hubv1.TokenLease) {
	consumedLeasesLock.Lock()
	consumedLeases = append(consumedLeases, lease.GetToken())
	consumedLeasesLock.Unlock()
	logging.Debug("consumed quota lease",
		"guard", guard,
		"feature", lease.GetFeature(),
		"weight", lease.GetWeight(),
		"priority_boost", lease.GetPriorityBoost())
}

func batchTokenConsumer() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ctx.Done():
			if global.QuotaServiceClient() != nil {
				// (attempt to) flush consumed token leases to hub when we exit
				ctx, cancel := context.WithTimeout(context.Background(), MAX_QUOTA_WAIT)
				defer cancel()
				global.QuotaServiceClient().SetTokenLeaseConsumed(
					metadata.NewOutgoingContext(ctx, global.XStanzaKey()),
					&hubv1.SetTokenLeaseConsumedRequest{Tokens: consumedLeases})
			}
			return
		case <-time.After(BATCH_TOKEN_CONSUME_INTERVAL):
			if global.QuotaServiceClient() != nil {
				consumedLeasesLock.Lock()
				if len(consumedLeases) == 0 {
					consumedLeasesLock.Unlock()
				} else {
					consumeTokenReq := &hubv1.SetTokenLeaseConsumedRequest{Tokens: consumedLeases}
					consumedLeases = []string{}
					consumedLeasesLock.Unlock()

					ctx, cancel := context.WithTimeout(context.Background(), MAX_QUOTA_WAIT)
					defer cancel()
					_, err := global.QuotaServiceClient().SetTokenLeaseConsumed(
						metadata.NewOutgoingContext(ctx, global.XStanzaKey()),
						consumeTokenReq)
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

func cachedLeaseManager() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(CACHED_LEASE_CHECK_INTERVAL):
			for guard := range cachedLeases {
				newCache := []*hubv1.TokenLease{}
				expiringLeaseCount := 0
				cachedLeasesLock[guard].Lock()
				cachedLeaseCount := len(cachedLeases[guard])

				// Check for and remove any expired leases
				for k, tl := range cachedLeases[guard] {
					if time.Now().Before(tl.GetExpiresAt().AsTime()) {
						newCache = append(newCache, cachedLeases[guard][k])
					} else {
						cachedLeasesUsed[guard] += 1
					}
				}

				// Check for number of leases within 2 seconds of expiring
				for _, tl := range newCache {
					if time.Now().Before(tl.GetExpiresAt().AsTime().Add(-2 * time.Second)) {
						expiringLeaseCount += 1
					}
				}

				// Add any additional leases waiting to be cached now
				waitingLeasesLock[guard].Lock()
				if len(waitingLeases[guard]) > 0 {
					newCache = append(newCache, waitingLeases[guard]...)
					cachedLeaseCount += len(waitingLeases[guard])
					cachedLeasesUsed[guard] = 0
					waitingLeases[guard] = []*hubv1.TokenLease{}
				}
				waitingLeasesLock[guard].Unlock()

				// Make a GetTokenLease request if >80% of our tokens are already used (or expiring soon)
				if global.QuotaServiceClient() != nil {
					if float32((cachedLeaseCount-expiringLeaseCount)/(cachedLeaseCount+cachedLeasesUsed[guard])) < 0.2 {
						go func() {
							ctx, cancel := context.WithTimeout(context.Background(), CACHED_LEASE_CHECK_INTERVAL)
							defer cancel()
							resp, err := global.QuotaServiceClient().GetTokenLease(
								metadata.NewOutgoingContext(ctx, global.XStanzaKey()),
								cachedLeasesReq[guard])
							if err != nil {
								logging.Error(err)
							}
							if len(resp.GetLeases()) > 0 {
								waitingLeasesLock[guard].Lock()
								waitingLeases[guard] = append(waitingLeases[guard], resp.GetLeases()...)
								waitingLeasesLock[guard].Unlock()
							}
						}()
					}
				}

				// Update the cached leases store
				cachedLeases[guard] = newCache
				cachedLeasesLock[guard].Unlock()
			}
		}
	}
}

func ValidateTokens(ctx context.Context, guard string, tokens []string) int {
	qsc := global.QuotaServiceClient()
	if qsc == nil {
		logging.Debug(
			"invalid quota service client, failing open",
			"count", atomic.AddInt64(&failOpenCount, 1),
		)
		return ValidateTokensFailOpen // fail open condition
	}
	gc := global.GetGuardConfig(ctx, guard)
	if gc == nil {
		logging.Debug(
			"invalid guard config, failing open",
			"count", atomic.AddInt64(&failOpenCount, 1),
		)
		return ValidateTokensFailOpen // fail open condition
	}

	if !gc.GetValidateIngressTokens() {
		return ValidateTokensSkipped // if we weren't asked to validate ingress tokens, don't
	}
	if len(tokens) == 0 {
		logging.Warn("validate ingress tokens was specified, but no tokens were found", "guard", guard)
		return ValidateTokensInvalid // fail fast in the case where we are supposed to validate, but no tokens found
	}

	gs := &hubv1.GuardSelector{Environment: global.GetServiceEnvironment(), Name: guard}
	vtr := &hubv1.ValidateTokenRequest{Tokens: tokenInfos(tokens, gs)}

	ctx, cancel := context.WithTimeout(context.Background(), MAX_QUOTA_WAIT)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			logging.Error(ctx.Err())
			return ValidateTokensFailOpen // deadline reached, log error and fail open
		default:
			resp, err := qsc.ValidateToken(metadata.NewOutgoingContext(ctx, global.XStanzaKey()), vtr)
			if err != nil {
				logging.Error(err)
				return ValidateTokensFailOpen // error from Stanza Hub, log error and fail open
			}
			for _, t := range resp.GetTokensValid() {
				if !t.Valid {
					return ValidateTokensInvalid
				}
			}
			return ValidateTokensValid
		}
	}
}

func tokenInfos(tokens []string, gs *hubv1.GuardSelector) (ti []*hubv1.TokenInfo) {
	for _, t := range tokens {
		ti = append(ti, &hubv1.TokenInfo{
			Token: t,
			Guard: gs,
		})
	}
	return ti
}
