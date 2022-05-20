package queue2

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/sinmetal/queue2/internal/trace"
)

type PubSubService struct {
	ps *pubsub.Client
}

func NewPubSubService(ctx context.Context, ps *pubsub.Client) (*PubSubService, error) {
	return &PubSubService{
		ps: ps,
	}, nil
}

func (s *PubSubService) PublishWithGet(ctx context.Context, projectID string, topicID string, msg *pubsub.Message) (serverID string, err error) {
	ctx = trace.StartSpan(ctx, "PubSubService/Publish")
	defer trace.EndSpan(ctx, err)

	topic := s.ps.TopicInProject(topicID, projectID)
	if msg.OrderingKey != "" {
		topic.EnableMessageOrdering = true
	}
	ret := topic.Publish(ctx, msg)
	serverID, err = ret.Get(ctx)
	if err != nil {
		return "", err
	}
	return serverID, nil
}

func (s *PubSubService) Publish(ctx context.Context, projectID string, topicID string, msg *pubsub.Message) (result *pubsub.PublishResult, err error) {
	ctx = trace.StartSpan(ctx, "PubSubService/Publish")
	defer trace.EndSpan(ctx, err)

	topic := s.ps.TopicInProject(topicID, projectID)
	if msg.OrderingKey != "" {
		topic.EnableMessageOrdering = true
	}
	ret := topic.Publish(ctx, msg)
	return ret, nil
}

func (s *PubSubService) Flush(ctx context.Context, projectID string, topicID string) {
	ctx = trace.StartSpan(ctx, "PubSubService/Flush")
	defer trace.EndSpan(ctx, nil)

	s.ps.TopicInProject(topicID, projectID).Flush()
}
