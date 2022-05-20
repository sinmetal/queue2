package queue2_test

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/sinmetal/queue2"
)

func TestPubSubService_Publish(t *testing.T) {
	ctx := context.Background()

	//pubSubClient, err := pubsub.NewClient(ctx, "sinmetal-queue2", option.WithEndpoint("asia-northeast1-pubsub.googleapis.com"))
	projectID := queue2.ProjectID()
	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		t.Fatal(err)
	}
	pubSubService, err := queue2.NewPubSubService(ctx, pubSubClient)
	if err != nil {
		t.Fatal(err)
	}

	attributes := map[string]string{"hello": "world"}
	{
		_, err = pubSubService.Publish(ctx, projectID, "hello", &pubsub.Message{
			Data:       []byte(time.Now().String()),
			Attributes: attributes,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		_, err = pubSubService.Publish(ctx, projectID, "hello-order", &pubsub.Message{
			Data:        []byte(time.Now().String()),
			Attributes:  attributes,
			OrderingKey: "key1",
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
