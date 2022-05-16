package queue2

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/vvakame/sdlog/aelog"
)

type HelloHandler struct {
	PubSubService *PubSubService
}

func NewHelloHandler(ctx context.Context, pubSubService *PubSubService) (*HelloHandler, error) {
	return &HelloHandler{
		PubSubService: pubSubService,
	}, nil
}

func (h *HelloHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fail := makeItFail()
	orderID := r.FormValue("order")
	{
		const topicID = "hello"
		attributes := map[string]string{"hello": "world"}
		attributes["fail"] = fail
		serverID, err := h.PubSubService.Publish(ctx, ProjectID(), topicID, &pubsub.Message{
			Data:       []byte(time.Now().String()),
			Attributes: attributes,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte(fmt.Sprintf("%s:%s\n", topicID, err.Error())))
			if err != nil {
				aelog.Errorf(ctx, "%s", err)
			}
			return
		}
		aelog.Infof(ctx, "Publish_ServerID:%s\n", serverID)
		_, err = w.Write([]byte(fmt.Sprintf("%s:%s\n", topicID, serverID)))
		if err != nil {
			aelog.Errorf(ctx, "%s", err)
		}
	}
	{
		const topicID = "hello-order"
		attributes := map[string]string{"hello": "world"}
		attributes["fail"] = fail
		serverID, err := h.PubSubService.Publish(ctx, ProjectID(), topicID, &pubsub.Message{
			Data:        []byte(time.Now().String()),
			Attributes:  attributes,
			OrderingKey: orderID,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte(fmt.Sprintf("%s:%s\n", topicID, err.Error())))
			if err != nil {
				aelog.Errorf(ctx, "%s", err)
			}
			return
		}
		aelog.Infof(ctx, "Publish_ServerID:%s\n", serverID)
		_, err = w.Write([]byte(fmt.Sprintf("%s:%s\n", topicID, serverID)))
		if err != nil {
			aelog.Errorf(ctx, "%s", err)
		}
	}

}

func makeItFail() string {
	if rand.Int() < 1000 {
		return "true"
	}
	return "false"
}
