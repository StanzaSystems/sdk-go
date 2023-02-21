package stanza

import (
	"net/http"

	httphandler "github.com/StanzaSystems/sdk-go/handlers/http"
)

func HttpGetHandler(url, decorator, feature string) (*http.Response, error) {
	return httphandler.HTTPGetHandler(url, decorator, feature)
}

func NewHttpInboundHandler(decorator string) (*httphandler.InboundHandler, error) {
	return httphandler.NewInboundHandler(gs.client.Name, decorator, SentinelEnabled())
}
