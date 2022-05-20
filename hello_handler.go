package queue2

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
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

	orderID := r.FormValue("order")
	workTimeSec := r.FormValue("workTimeSec")
	baseAttr := map[string]string{"hello": "world"}
	baseAttr["workTimeSec"] = workTimeSec
	{
		attr := map[string]string{}
		for k, v := range baseAttr {
			attr[k] = v
		}

		fail := makeItFail()
		attr["fail"] = fail
		const topicID = "hello"
		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		serverID, err := h.PubSubService.PublishWithGet(ctx, ProjectID(), topicID, &pubsub.Message{
			Data:       []byte(time.Now().String()),
			Attributes: attr,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte(fmt.Sprintf("%s:%s\n", topicID, err.Error())))
			if err != nil {
				aelog.Errorf(ctx, "%s\n", err)
			}
			return
		}
		aelog.Infof(ctx, "Publish_ServerID:%s:%s\n", serverID, fail)
		_, err = w.Write([]byte(fmt.Sprintf("%s:%s\n", topicID, serverID)))
		if err != nil {
			aelog.Errorf(ctx, "%s\n", err)
		}
	}
	{
		const topicID = "hello-order"
		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		attr := map[string]string{}
		for k, v := range baseAttr {
			attr[k] = v
		}
		failPublishNumber := []string{}
		for i := 0; i < 10; i++ {
			fail := makeItFail()
			attr["fail"] = fail
			pn := fmt.Sprintf("%03d", i)
			attr["PublishNumber"] = pn
			if strings.ToLower(fail) == "true" {
				failPublishNumber = append(failPublishNumber, pn)
			}
			_, err := h.PubSubService.Publish(ctx, ProjectID(), topicID, &pubsub.Message{
				Data:        []byte(time.Now().String()),
				Attributes:  attr,
				OrderingKey: orderID,
			})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte(fmt.Sprintf("%s:%s\n", topicID, err.Error())))
				if err != nil {
					aelog.Errorf(ctx, "%s\n", err)
				}
				return
			}
		}

		h.PubSubService.Flush(ctx, ProjectID(), topicID)
		_, err := w.Write([]byte(fmt.Sprintf("Flush %s.FailPublishNumbers:%s\n", topicID, strings.Join(failPublishNumber, ","))))
		if err != nil {
			aelog.Errorf(ctx, "%s\n", err)
		}
		return
	}
}

func makeItFail() string {
	if rand.Int() < 1000 {
		return "true"
	}
	return "false"
}
