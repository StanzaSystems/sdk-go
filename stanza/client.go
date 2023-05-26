package stanza

import (
	"context"

	"github.com/StanzaSystems/sdk-go/logging"
	v1 "github.com/StanzaSystems/sdk-go/proto/stanza/hub/v1"
	"google.golang.org/grpc"
)

func GetBearerToken(host string, apiKey string) (string, error) {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		logging.Error(err, "msg", "failed to connect to stanza hub")
	}
	defer conn.Close()

	client := v1.NewAuthServiceClient(conn)
	req := &v1.GetBearerTokenRequest{ApiKey: apiKey}
	res, err := client.GetBearerToken(context.Background(), req)
	if err != nil {
		logging.Error(err, "msg", "failed to to get bearer token")
	}
	token := res.GetBearerToken()

	return token, err
}
