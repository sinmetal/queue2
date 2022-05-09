package queue2

import (
	"context"
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

	serverID, err := h.PubSubService.Publish(ctx, ProjectID(), "hello", &pubsub.Message{
		Data:       []byte(time.Now().String()),
		Attributes: nil,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			aelog.Errorf(ctx, "%s", err)
		}
		return
	}
	aelog.Infof(ctx, "Publish_ServerID:%s\n", serverID)
	_, err = w.Write([]byte(serverID))
	if err != nil {
		aelog.Errorf(ctx, "%s", err)
	}
}
