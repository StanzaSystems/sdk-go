package stanza

import (
	"context"

	"github.com/StanzaSystems/sdk-go/logging"
	v1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
)

func GetBearerToken(host string, apiKey string) (string, error) {
	client := v1.NewAuthServiceClient()
	req := &v1.GetBearerTokenRequest{ApiKey: apiKey}
	res, err := client.GetBearerToken(context.Background(), req)
	if err != nil {
		logging.Error(err, "msg", "failed to to get bearer token")
	}
	token := res.GetBearerToken()

	return token, err
}
