package tasks2

import (
	"context"

	"github.com/sinmetal/queue2/internal/trace"
	cloudtasksbox "github.com/sinmetalcraft/gcpbox/cloudtasks"
)

type TasksService struct {
	queue    *cloudtasksbox.Queue
	tasksbox *cloudtasksbox.Service
}

func NewTasksService(ctx context.Context, queue *cloudtasksbox.Queue, service *cloudtasksbox.Service) (*TasksService, error) {
	return &TasksService{
		queue,
		service,
	}, nil
}

func (s *TasksService) CreateGetTask(ctx context.Context, task *cloudtasksbox.GetTask) (taskName string, err error) {
	ctx = trace.StartSpan(ctx, "TasksService/CreateGetTask")
	defer trace.EndSpan(ctx, err)

	tn, err := s.tasksbox.CreateGetTask(ctx, s.queue, task)
	if err != nil {
		return "", err
	}
	return tn, nil
}
