package handlers

import (
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
)

type InboundHandler struct {
	*Handler
	tlr map[string]*hubv1.GetTokenLeaseRequest
}

// NewInboundHandler returns a new InboundHandler
func NewInboundHandler(apikey, clientId, environment, service string, otelEnabled, sentinelEnabled bool) (*InboundHandler, error) {
	h, err := NewHandler(apikey, clientId, environment, service, otelEnabled, sentinelEnabled)
	return &InboundHandler{h, make(map[string]*hubv1.GetTokenLeaseRequest)}, err
}

func (h *InboundHandler) SetTokenLeaseRequest(d string, tlr *hubv1.GetTokenLeaseRequest) {
	tlr.ClientId = &h.clientId
	tlr.Selector.Environment = h.environment
	if h.tlr[d] == nil {
		h.tlr[d] = tlr
	}
}

func (h *InboundHandler) TokenLeaseRequest(dec string) *hubv1.GetTokenLeaseRequest {
	return h.tlr[dec]
}
