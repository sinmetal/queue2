package queue2

import (
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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

	publishNumber := pubSubBody.Message.Attributes["PublishNumber"]
	line := struct {
		Subscription     string
		ReceiveMessageID string
		OrderingKey      string
		PublishNumber    string
		PublishTime      int64
		ReceiveLocalTime int64
		DeliveryAttempt  int
	}{
		pubSubBody.Subscription,
		pubSubBody.Message.MessageID,
		pubSubBody.Message.OrderingKey,
		publishNumber,
		pubSubBody.Message.PublishTime.UnixMicro(),
		time.Now().UnixMicro(),
		pubSubBody.Message.DeliveryAttempt,
	}
	lineJ, err := json.Marshal(line)
	if err != nil {
		aelog.Errorf(ctx, "failed json.Unmarshal %s", err)
		return
	}
	aelog.Infof(ctx, `__RECEIVE_MESSAGE__:%s`, lineJ)

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

	if strings.HasSuffix(pubSubBody.Subscription, "dead-letter") && rand.Intn(100) < 10 {
		w.WriteHeader(http.StatusOK)
		return
	}

	fail := false
	failValue, ok := pubSubBody.Message.Attributes["fail"]
	if ok {
		if strings.ToLower(failValue) == "true" {
			fail = true
		}
	}

	v, ok := pubSubBody.Message.Attributes["workTimeSec"]
	if ok && v != "" {
		workTimeSec, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			aelog.Errorf(ctx, "invalid workTimeSec format.value=%s", v)
			return
		}
		aelog.Infof(ctx, "sleep:%d sec", workTimeSec)
		time.Sleep(time.Duration(workTimeSec) * time.Second)
	}
	if fail {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
