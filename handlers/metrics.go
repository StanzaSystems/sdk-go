package handlers

import (
	"go.opentelemetry.io/otel/attribute"
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
)

var (
	clientIdKey    = attribute.Key("client_id")
	customerIdKey  = attribute.Key("customer_id")
	guardKey       = attribute.Key("guard")
	environmentKey = attribute.Key("environment")
	featureKey     = attribute.Key("feature")
	serviceKey     = attribute.Key("service")
	reasonKey      = attribute.Key("reason")
)

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
