package handlers

type OutboundHandler struct {
	*Handler
}

func NewOutboundHandler(gn string, fn *string, pb *int32, dw *float32, kv *map[string]string) (*OutboundHandler, error) {
	h, err := NewHandler(gn, fn, pb, dw, kv)
	return &OutboundHandler{h}, err
}
