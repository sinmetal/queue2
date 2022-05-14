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

func (s *PubSubService) Publish(ctx context.Context, projectID string, topicID string, msg *pubsub.Message) (serverID string, err error) {
	ctx = trace.StartSpan(ctx, "PubSubService/Publish")
	defer trace.EndSpan(ctx, err)

	ret := s.ps.TopicInProject(topicID, projectID).Publish(ctx, msg)
	serverID, err = ret.Get(ctx)
	if err != nil {
		return "", err
	}
	return serverID, nil
}
