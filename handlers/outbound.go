package handlers

type OutboundHandler struct {
	*Handler
}

func NewOutboundHandler(apikey, clientId, environment, service string, otelEnabled, sentinelEnabled bool, instrumentationName string, instrumentationVersion string) *OutboundHandler {
	return &OutboundHandler{NewHandler(apikey, clientId, environment, service, otelEnabled, sentinelEnabled, instrumentationName, instrumentationVersion)}
}
