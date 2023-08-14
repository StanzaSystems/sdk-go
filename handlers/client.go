package handlers

type OutboundHandler struct {
	*Handler
}

func NewOutboundHandler() (*OutboundHandler, error) {
	h, err := NewHandler()
	return &OutboundHandler{h}, err
}
