package queue2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/vvakame/sdlog/aelog"
)

type ReceiveHandler struct {
}

type Body struct {
	Message      *Message `json:"message"`
	Subscription string   `json:"subscription"`
}

type Message struct {
	// ID identifies this message. This ID is assigned by the server and is
	// populated for Messages obtained from a subscription.
	//
	// This field is read-only.
	MessageID string `json:"messageId"`

	// Data is the actual data in the message.
	Data []byte `json:"data"`

	// Attributes represents the key-value pairs the current message is
	// labelled with.
	Attributes map[string]string `json:"attributes"`

	// PublishTime is the time at which the message was published. This is
	// populated by the server for Messages obtained from a subscription.
	//
	// This field is read-only.
	PublishTime time.Time `json:"publishTime"`

	// DeliveryAttempt is the number of times a message has been delivered.
	// This is part of the dead lettering feature that forwards messages that
	// fail to be processed (from nack/ack deadline timeout) to a dead letter topic.
	// If dead lettering is enabled, this will be set on all attempts, starting
	// with value 1. Otherwise, the value will be nil.
	// This field is read-only.
	DeliveryAttempt int `json:"deliveryAttempt"`

	// OrderingKey identifies related messages for which publish order should
	// be respected. If empty string is used, message will be sent unordered.
	OrderingKey string `json:"orderingKey"`
}

func NewReceiveHandler(ctx context.Context) (*ReceiveHandler, error) {
	return &ReceiveHandler{}, nil
}

func (h *ReceiveHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			aelog.Errorf(ctx, "failed read body %s", err)
		}
		return
	}
	aelog.Infof(ctx, "rawBody:%s\n", string(body))

	pubSubBody := &Body{}
	if err := json.Unmarshal(body, pubSubBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			aelog.Errorf(ctx, "failed json.Unmarshal %s", err)
		}
		return
	}
	aelog.Infof(ctx, `__RECEIVE_MESSAGE__:{"Subscription":%s,"ReceiveMessageID":%s,"PublishTime":%d}\n`,
		pubSubBody.Subscription, pubSubBody.Message.MessageID, pubSubBody.Message.PublishTime.UnixMicro())

	j, err := json.Marshal(pubSubBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			aelog.Errorf(ctx, "failed json.Marshal %s", err)
		}
		return
	}
	aelog.Infof(ctx, "pubSubBody:%s\n", string(j))
}
