package rabbitmq

import (
	"gitlab.thovnn.vn/core/golang-sdk/spubsub"
	"google.golang.org/grpc"
)

func NewPubSubClient(url string) spubsub.PubsubClient {
	connection, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := spubsub.NewPubsubClient(connection)
	return client
}
