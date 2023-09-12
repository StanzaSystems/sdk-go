package handlers

type OutboundHandler struct {
	*Handler
}

func NewOutboundHandler(gn string, fn *string, pb *int32, dw *float32) (*OutboundHandler, error) {
	h, err := NewHandler(gn, fn, pb, dw)
	return &OutboundHandler{h}, err
}
