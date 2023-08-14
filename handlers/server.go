package handlers

import (
	hubv1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
	"google.golang.org/protobuf/proto"
)

type InboundHandler struct {
	*Handler
	tlr map[string]*hubv1.GetTokenLeaseRequest
}

// NewInboundHandler returns a new InboundHandler
func NewInboundHandler() (*InboundHandler, error) {
	h, err := NewHandler()
	return &InboundHandler{h, make(map[string]*hubv1.GetTokenLeaseRequest)}, err
}

func (h *InboundHandler) SetTokenLeaseRequest(d string, tlr *hubv1.GetTokenLeaseRequest) {
	tlr.ClientId = proto.String(h.ClientID())
	tlr.Selector.Environment = h.Environment()

	// TODO LOCK
	if h.tlr[d] == nil {
		h.tlr[d] = tlr
	}
}

func (h *InboundHandler) TokenLeaseRequest(dec string) *hubv1.GetTokenLeaseRequest {
	// TODO LOCK
	return h.tlr[dec]
}
