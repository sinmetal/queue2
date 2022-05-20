package queue2

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/sinmetal/queue2/internal/trace"
)

type PubSubService struct {
	ps    *pubsub.Client
	topic *pubsub.Topic
}

func NewPubSubService(ctx context.Context, ps *pubsub.Client, topicID string, projectID string, enableMessageOrdering bool) (*PubSubService, error) {
	topic := ps.TopicInProject(topicID, projectID)
	topic.EnableMessageOrdering = enableMessageOrdering
	return &PubSubService{
		topic: topic,
		ps:    ps,
	}, nil
}

func (s *PubSubService) PublishWithGet(ctx context.Context, msg *pubsub.Message) (serverID string, err error) {
	ctx = trace.StartSpan(ctx, "PubSubService/Publish")
	defer trace.EndSpan(ctx, err)

	ret := s.Publish(ctx, msg)
	serverID, err = ret.Get(ctx)
	if err != nil {
		return "", err
	}
	return serverID, nil
}

func (s *PubSubService) Publish(ctx context.Context, msg *pubsub.Message) *pubsub.PublishResult {
	ctx = trace.StartSpan(ctx, "PubSubService/Publish")
	defer trace.EndSpan(ctx, nil)

	return s.topic.Publish(ctx, msg)
}

func (s *PubSubService) Flush(ctx context.Context) {
	ctx = trace.StartSpan(ctx, "PubSubService/Flush")
	defer trace.EndSpan(ctx, nil)

	s.topic.Flush()
}
