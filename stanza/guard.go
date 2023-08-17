package stanza

import (
	"context"
	"fmt"

	"github.com/StanzaSystems/sdk-go/handlers"
	"github.com/StanzaSystems/sdk-go/logging"
)

var (
	h *handlers.Handler = nil
)

func Guard(ctx context.Context, name string) *handlers.Guard {
	if h == nil {
		var err error
		h, err = handlers.NewHandler()
		if err != nil {
			err = fmt.Errorf("failed to create guard handler: %s", err)
			logging.Error(err)
			return h.NewGuardError(err)
		}
	}
	return h.NewGuard(ctx, name)
}
