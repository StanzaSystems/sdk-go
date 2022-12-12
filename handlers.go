package stanza

import (
	"errors"
	"net/http"

	"github.com/StanzaSystems/sdk-go/logging"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

func HttpInboundHandler(resource string, request *http.Request) int {
	// Wrap OTEL (https://github.com/gofiber/contrib/tree/main/otelfiber)

	entry, sentinelErr := api.Entry(resource, api.WithTrafficType(base.Inbound), api.WithResourceType(base.ResTypeWeb))
	if sentinelErr != nil {
		// capture metrics
		// blocked count
		// blocked count by sentinel block type?
		// total count
		// latency percentiles?

		// what do we want from that http request?
		// I think potentially a lot for our trace/span...
		// not sure about metrics though -- maybe some "path" based counts?

		logging.Error(
			errors.New("SentinelBlockError"), sentinelErr.BlockMsg(),
			"SentinelBlockType", sentinelErr.BlockType().String(),
			"SentinelBlockValue", sentinelErr.TriggeredValue(),
		)
		logging.Debug("", "SentinelBlockRule", sentinelErr.TriggeredRule().String())

		// TODO: allow sentinel "customize block fallback" to override this 429 default
		return http.StatusTooManyRequests
	} else {
		// Passed, wrap the logic here.
		logging.Debug("passed")

		// capture metrics
		// passed count
		// total count
		// latency percentiles?

		// Be sure the entry is exited finally.
		entry.Exit()
		return http.StatusOK
	}
}
