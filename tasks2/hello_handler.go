package tasks2

import (
	"context"
	"fmt"
	"net/http"
	"time"

	cloudtasksbox "github.com/sinmetalcraft/gcpbox/cloudtasks"
	"github.com/vvakame/sdlog/aelog"
)

type HelloHandler struct {
	helloQueueTasksService *TasksService
}

func NewHelloHandler(ctx context.Context, helloQueueTasksService *TasksService) (*HelloHandler, error) {
	return &HelloHandler{
		helloQueueTasksService: helloQueueTasksService,
	}, nil
}

func (h *HelloHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	forceFail := r.FormValue("forceFail")
	req, err := http.NewRequest(http.MethodGet, "https://queue2-b2ikiuzo3a-an.a.run.app/tasks/receive", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(fmt.Sprintf("failed http.NewRequest %s\n", err.Error())))
		if err != nil {
			aelog.Errorf(ctx, "%s\n", err)
		}
		return
	}

	params := req.URL.Query()
	params.Add("forceFail", forceFail)
	params.Add("publishTime", fmt.Sprintf("%d", time.Now().UnixMicro()))
	req.URL.RawQuery = params.Encode()

	tn, err := h.helloQueueTasksService.CreateGetTask(ctx, &cloudtasksbox.GetTask{
		Audience:     req.URL.String(),
		Headers:      nil,
		RelativeURI:  req.URL.String(),
		ScheduleTime: time.Time{},
		Deadline:     0,
		Name:         "",
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
		if err != nil {
			aelog.Errorf(ctx, "%s\n", err)
		}
		return
	}
	aelog.Infof(ctx, "Publish_TaskName:%s:%s\n", tn, forceFail)
	_, err = w.Write([]byte(fmt.Sprintf("%s\n", tn)))
	if err != nil {
		aelog.Errorf(ctx, "%s\n", err)
	}
}
