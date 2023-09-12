package handlers

type InboundHandler struct {
	*Handler
}

// NewInboundHandler returns a new InboundHandler
func NewInboundHandler(gn string, fn *string, pb *int32, dw *float32) (*InboundHandler, error) {
	h, err := NewHandler(gn, fn, pb, dw)
	return &InboundHandler{h}, err
}
