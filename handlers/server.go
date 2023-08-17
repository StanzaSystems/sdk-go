package handlers

import (
	"sync"

	hubv1 "github.com/StanzaSystems/sdk-go/gen/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/hub"
	"google.golang.org/protobuf/proto"
)

type InboundHandler struct {
	*Handler
	lock *sync.RWMutex
	tlr  map[string]*hubv1.GetTokenLeaseRequest
}

// NewInboundHandler returns a new InboundHandler
func NewInboundHandler() (*InboundHandler, error) {
	h, err := NewHandler()
	return &InboundHandler{h, &sync.RWMutex{}, make(map[string]*hubv1.GetTokenLeaseRequest)}, err
}

func (h *InboundHandler) SetTokenLeaseRequest(d string, tlr *hubv1.GetTokenLeaseRequest) {
	tlr.ClientId = proto.String(h.ClientID())
	tlr.Selector.Environment = h.Environment()
	if h.tlr[d] == nil {
		h.lock.Lock()
		h.tlr[d] = tlr
		h.lock.Unlock()
	}
}

func (h *InboundHandler) TokenLeaseRequest(guard string) *hubv1.GetTokenLeaseRequest {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if tlr, ok := h.tlr[guard]; ok {
		return tlr
	}
	return hub.NewTokenLeaseRequest(guard)
}
