package handlers

type OutboundHandler struct {
	*Handler
}

func NewOutboundHandler(apikey, clientId, environment, service string, otelEnabled, sentinelEnabled bool) (*OutboundHandler, error) {
	h, err := NewHandler(apikey, clientId, environment, service, otelEnabled, sentinelEnabled)
	return &OutboundHandler{h}, err
}
