package handlers

type InboundHandler struct {
	*Handler
}

// NewInboundHandler returns a new InboundHandler
func NewInboundHandler() (*InboundHandler, error) {
	h, err := NewHandler()
	return &InboundHandler{h}, err
}
