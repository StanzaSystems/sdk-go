package hub

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	hubv1 "buf.build/gen/go/stanza/apis/protocolbuffers/go/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/global"
	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/StanzaSystems/sdk-go/otel"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	MAX_QUOTA_WAIT               = 1 * time.Second
	CACHED_LEASE_CHECK_INTERVAL  = 200 * time.Millisecond // TODO: what should this be set to?
	BATCH_TOKEN_CONSUME_INTERVAL = 200 * time.Millisecond // TODO: what should this be set to?
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

func NewTokenLeaseRequest(ctx context.Context, gn string, fn *string, pb *int32, dw *float32, tags *map[string]string) (context.Context, *hubv1.GetTokenLeaseRequest) {
	tlr := hubv1.GetTokenLeaseRequest{
		ClientId: proto.String(global.GetClientID()),
		Selector: &hubv1.GuardFeatureSelector{
			GuardName:   gn,
			Environment: global.GetServiceEnvironment(),
		},
	}

	// Inspect Baggage and Headers for Feature, propagate through context if found
	ctx, feat := otel.GetFeature(ctx, fn)
	if feat != nil {
		tlr.Selector.FeatureName = feat
	}

	// Inspect Baggage and Headers for PriorityBoost, propagate through context if found
	ctx, boost := otel.GetPriorityBoost(ctx, pb)
	if boost != nil {
		tlr.PriorityBoost = boost
	}

	// DefaultWeight can not be set via Baggage or Headers
	if dw != nil {
		tlr.DefaultWeight = dw
	}

	// Add Tags (if Guard config allows it)
	if tags != nil {
		if len(*tags) > 0 {
			guardConfig, _, err := global.GetGuardConfig(ctx, gn)
			if err != nil {
				logging.Error(err)
			} else {
				for k, v := range *tags {
					if slices.Contains(guardConfig.QuotaTags, k) {
						tlr.Selector.Tags = append(tlr.Selector.Tags, &hubv1.Tag{Key: k, Value: v})
					} else {
						logging.Info("skipping unknown tag", "tag", k, "guard", gn)
					}
				}
			}
		}
	}
	return ctx, &tlr
}

func CheckQuota(_ context.Context, tlr *hubv1.GetTokenLeaseRequest) (hubv1.Quota, string, error) {
	if tlr == nil || tlr.Selector == nil {
		errMsg := "invalid token lease request, failing open"
		logging.Debug(errMsg, "count", atomic.AddInt64(&failOpenCount, 1))
		return hubv1.Quota_QUOTA_NOT_EVAL, "", errors.New(errMsg)
	}
	qsc := global.QuotaServiceClient()
	if qsc == nil {
		errMsg := "invalid quota service client, failing open"
		logging.Debug(errMsg, "count", atomic.AddInt64(&failOpenCount, 1))
		return hubv1.Quota_QUOTA_NOT_EVAL, "", errors.New(errMsg)
	}
	guard := tlr.GetSelector().GetGuardName()

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
							return hubv1.Quota_QUOTA_GRANTED, tl.Token, nil
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

	for {
		select {
		case <-ctx.Done():
			return hubv1.Quota_QUOTA_TIMEOUT, "", ctx.Err() // deadline reached, log error and fail open
		default:
			resp, err := qsc.GetTokenLease(metadata.NewOutgoingContext(ctx, global.XStanzaKey()), tlr)
			if err != nil {
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
				return hubv1.Quota_QUOTA_ERROR, "", err // just fail open (for now)
			}
			leases := resp.GetLeases()
			if len(leases) == 0 {
				return hubv1.Quota_QUOTA_BLOCKED, "", nil // not an error, there were no leases available
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
			return hubv1.Quota_QUOTA_GRANTED, leases[0].Token, nil
		}
	}
}

func consumeLease(guard string, lease *hubv1.TokenLease) {
	consumedLeasesLock.Lock()
	consumedLeases = append(consumedLeases, lease.GetToken())
	consumedLeasesLock.Unlock()
	// TODO: Fix hub bug (feature, weight, and priority_boost aren't optional)
	// logging.Debug("consumed quota lease",
	// 	"guard", guard,
	// 	"feature", lease.Feature,
	// 	"weight", lease.Weight,
	// 	"priority_boost", lease.PriorityBoost)
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
					&hubv1.SetTokenLeaseConsumedRequest{
						Tokens:      consumedLeases,
						Environment: global.GetServiceEnvironment(),
					})
			}
			return
		case <-time.After(BATCH_TOKEN_CONSUME_INTERVAL):
			if global.QuotaServiceClient() != nil {
				consumedLeasesLock.Lock()
				if len(consumedLeases) == 0 {
					consumedLeasesLock.Unlock()
				} else {
					consumeTokenReq := &hubv1.SetTokenLeaseConsumedRequest{
						Tokens:      consumedLeases,
						Environment: global.GetServiceEnvironment(),
					}
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

func ValidateTokens(_ context.Context, guard string, tokens []string) (hubv1.Token, error) {
	qsc := global.QuotaServiceClient()
	if qsc == nil {
		errMsg := "invalid quota service client, failing open"
		logging.Debug(errMsg,
			"guard", guard,
			"count", atomic.AddInt64(&failOpenCount, 1))
		return hubv1.Token_TOKEN_NOT_EVAL, errors.New(errMsg)
	}

	if len(tokens) == 0 {
		// fail fast in the case where we are supposed to validate, but no tokens found
		logging.Warn("validate ingress tokens was specified, but no tokens were found", "guard", guard)
		return hubv1.Token_TOKEN_NOT_VALID, nil
	}

	gs := &hubv1.GuardSelector{Environment: global.GetServiceEnvironment(), Name: guard}
	vtr := &hubv1.ValidateTokenRequest{Tokens: tokenInfos(tokens, gs)}

	ctx, cancel := context.WithTimeout(context.Background(), MAX_QUOTA_WAIT)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return hubv1.Token_TOKEN_VALIDATION_TIMEOUT, ctx.Err() // deadline reached, log error and fail open
		default:
			resp, err := qsc.ValidateToken(metadata.NewOutgoingContext(ctx, global.XStanzaKey()), vtr)
			if err != nil {
				return hubv1.Token_TOKEN_VALIDATION_ERROR, err // error from Stanza Hub, log error and fail open
			}
			for _, t := range resp.GetTokensValid() {
				if !t.Valid {
					return hubv1.Token_TOKEN_NOT_VALID, nil
				}
			}
			return hubv1.Token_TOKEN_VALID, nil
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
