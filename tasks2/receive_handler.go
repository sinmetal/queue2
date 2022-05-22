package tasks2

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	cloudtasksbox "github.com/sinmetalcraft/gcpbox/cloudtasks"
	"github.com/vvakame/sdlog/aelog"
)

type ReceiveHandler struct {
}

func NewReceiveHandler(ctx context.Context) (*ReceiveHandler, error) {
	return &ReceiveHandler{}, nil
}

func (h *ReceiveHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tasksHeader, err := cloudtasksbox.GetHeader(r)
	if err != nil {
		aelog.Errorf(ctx, "failed cloudtasksbox.GetHeader %s", err)
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			aelog.Errorf(ctx, "failed write body %s", err)
		}
		return
	}

	publishTime := r.FormValue("publishTime")
	publishTimeMicroSec, err := strconv.ParseInt(publishTime, 10, 64)
	if err != nil {
		aelog.Infof(ctx, "invalid format publishTime. value=%s\n", publishTime)
	}

	line := struct {
		Header      *cloudtasksbox.Header
		PublishTime int64
	}{
		tasksHeader,
		publishTimeMicroSec,
	}
	lineJ, err := json.Marshal(line)
	if err != nil {
		aelog.Errorf(ctx, "failed json.Marshal %s", err)
		return
	}
	aelog.Infof(ctx, `__RECEIVE_TASK__:%s`, lineJ)
}
