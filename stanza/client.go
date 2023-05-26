package stanza

import (
	"context"
	"errors"
	"net/http"

	v1 "github.com/StanzaSystems/sdk-go/gen/go/stanza/hub/v1"
	"github.com/StanzaSystems/sdk-go/gen/go/stanza/hub/v1/hubv1connect"
	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/bufbuild/connect-go"
)

func GetBearerToken(host string, apiKey string) (string, error) {
	client := hubv1connect.NewAuthServiceClient(
		http.DefaultClient,
		host,
	)
	req := connect.NewRequest(&v1.GetBearerTokenRequest{ApiKey: apiKey})
	req.Header().Set("x-stanza-key", apiKey)
	res, err := client.GetBearerToken(context.Background(), req)
	if err != nil {
		logging.Error(err, "msg", "failed to to get bearer token")
	}
	token := res.Msg.GetBearerToken()
	err = errors.New(res.Msg.GetErrorMessage())

	return token, err
}
