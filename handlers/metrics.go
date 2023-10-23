package handlers

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	localReason = "local_reason"
	tokenReason = "token_reason"
	quotaReason = "quota_reason"
)

var (
	clientIdKey    = attribute.Key("client_id")
	customerIdKey  = attribute.Key("customer_id")
	guardKey       = attribute.Key("guard")
	environmentKey = attribute.Key("environment")
	featureKey     = attribute.Key("feature")
	serviceKey     = attribute.Key("service")
	errorKey       = attribute.Key("error")
	localReasonKey = attribute.Key(localReason)
	tokenReasonKey = attribute.Key(tokenReason)
	quotaReasonKey = attribute.Key(quotaReason)
)
