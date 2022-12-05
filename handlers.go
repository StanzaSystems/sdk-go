package stanza

import (
	"net/http"

	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

func HttpInboundHandler(resource string, request *http.Request) error {
	// Wrap OTEL (https://github.com/gofiber/contrib/tree/main/otelfiber)

	e, b := api.Entry(resource, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
	if b != nil {
		// Blocked, wrap the logic here.
		// We could get the block reason from the BlockError.
		logging.Debug("blocked")

		// Be sure the entry is exited finally.
		e.Exit()
	} else {
		// Passed, wrap the logic here.
		logging.Debug("passed")

		// Be sure the entry is exited finally.
		e.Exit()
	}
	return nil
}
