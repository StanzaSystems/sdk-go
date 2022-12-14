package otel

import (
	"context"

	"github.com/StanzaSystems/sdk-go/global"
)

func Init(ctx context.Context) error {
	if err := global.SetOtelResource(ctx); err != nil {
		return err
	}
	return nil
}
