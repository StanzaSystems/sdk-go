package handlers

import (
	"github.com/StanzaSystems/sdk-go/global"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	ReasonUnknown = iota
	ReasonFailOpen
	ReasonDarkLaunch
	ReasonQuota
	ReasonQuotaToken
	ReasonQuotaFailOpen
	ReasonQuotaCheckDisabled
	ReasonQuotaInvalidToken
	ReasonQuotaUnknown
	ReasonSentinel

	// Stanza SDK Metrics:
	// https://github.com/StanzaSystems/sdk-spec#telemetry-metrics
	stanzaAllowed         = "stanza.guard.allowed"          // counter
	stanzaAllowedSuccess  = "stanza.guard.allowed.success"  // counter
	stanzaAllowedFailure  = "stanza.guard.allowed.failure"  // counter
	stanzaAllowedUnknown  = "stanza.guard.allowed.unknown"  // counter
	stanzaAllowedDuration = "stanza.guard.allowed.duration" // histogram (milliseconds)
	stanzaBlocked         = "stanza.guard.blocked"          // counter
)

var (
	clientIdKey    = attribute.Key("client_id")
	customerIdKey  = attribute.Key("customer_id")
	guardKey       = attribute.Key("guard")
	environmentKey = attribute.Key("environment")
	featureKey     = attribute.Key("feature")
	serviceKey     = attribute.Key("service")
	reasonKey      = attribute.Key("reason")

	stanzaMeter *meter
)

type meter struct {
	AllowedCount        metric.Int64Counter
	AllowedSuccessCount metric.Int64Counter
	AllowedFailureCount metric.Int64Counter
	AllowedUnknownCount metric.Int64Counter
	AllowedDuration     metric.Float64Histogram
	BlockedCount        metric.Int64Counter
}

func reason(reason int) attribute.KeyValue {
	switch reason {
	case ReasonFailOpen:
		return reasonKey.String("fail_open")
	case ReasonDarkLaunch:
		return reasonKey.String("dark_launch")
	case ReasonQuota:
		return reasonKey.String("quota")
	case ReasonQuotaToken:
		return reasonKey.String("quota_token")
	case ReasonQuotaFailOpen:
		return reasonKey.String("quota_fail_open")
	case ReasonQuotaCheckDisabled:
		return reasonKey.String("quota_check_disabled")
	case ReasonQuotaInvalidToken:
		return reasonKey.String("quota_invalid_token")
	case ReasonQuotaUnknown:
		return reasonKey.String("quota_unknown")
	}
	return reasonKey.String("unknown")
}

func GetStanzaMeter() (*meter, error) {
	if stanzaMeter != nil {
		return stanzaMeter, nil
	}
	om := otel.Meter(
		global.InstrumentationName(),
		global.InstrumentationMetricVersion(),
	)

	var err error
	var m meter
	m.AllowedCount, err = om.Int64Counter(
		stanzaAllowed,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of executions that were allowed"))
	if err != nil {
		return nil, err
	}
	m.BlockedCount, err = om.Int64Counter(
		stanzaBlocked,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of executions that were backpressured"))
	if err != nil {
		return nil, err
	}
	m.AllowedSuccessCount, err = om.Int64Counter(
		stanzaAllowedSuccess,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of executions that succeeded"))
	if err != nil {
		return nil, err
	}
	m.AllowedFailureCount, err = om.Int64Counter(
		stanzaAllowedFailure,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of executions that failed"))
	if err != nil {
		return nil, err
	}
	m.AllowedUnknownCount, err = om.Int64Counter(
		stanzaAllowedUnknown,
		metric.WithUnit("1"),
		metric.WithDescription("measures the number of executions where the success (or failure) was unknown"))
	if err != nil {
		return nil, err
	}
	m.AllowedDuration, err = om.Float64Histogram(
		stanzaAllowedDuration,
		metric.WithUnit("ms"),
		metric.WithDescription("measures the total executions time of guarded requests"))
	if err != nil {
		return nil, err
	}

	stanzaMeter = &m
	return stanzaMeter, nil
}
