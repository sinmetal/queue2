package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/pubsub"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/sinmetal/queue2"
	"github.com/sinmetal/queue2/pubsub2"
	"github.com/sinmetal/queue2/tasks2"
	cloudtasksbox "github.com/sinmetalcraft/gcpbox/cloudtasks"
	metadatabox "github.com/sinmetalcraft/gcpbox/metadata"
	"github.com/vvakame/sdlog/aelog"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

type Handlers struct {
	PubSubHelloHandler   *pubsub2.HelloHandler
	PubSubReceiveHandler *pubsub2.ReceiveHandler
	TasksHelloHandler    *tasks2.HelloHandler
	TasksReceiveHandler  *tasks2.ReceiveHandler
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, err := fmt.Fprintf(w, "Hello, Queue2")
	if err != nil {
		aelog.Errorf(ctx, "err=%+\nv", err)
	}
}

func main() {
	ctx := context.Background()

	pID := queue2.ProjectID()
	serviceAccountEmail := queue2.ServiceAccountEmail()
	fmt.Printf("ProjectID is %s\n", pID)
	if metadatabox.OnGCP() {
		exporter, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID: pID,
		})
		if err != nil {
			panic(err)
		}
		trace.RegisterExporter(exporter)
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
		fmt.Println("start Cloud Trace")
	}

	handlers, err := createHandlers(ctx, pID, serviceAccountEmail)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/pubsub/hello", handlers.PubSubHelloHandler.Handle)
	mux.HandleFunc("/pubsub/receive", handlers.PubSubReceiveHandler.Handle)
	mux.HandleFunc("/tasks/hello", handlers.TasksHelloHandler.Handle)
	mux.HandleFunc("/tasks/receive", handlers.TasksReceiveHandler.Handle)
	mux.HandleFunc("/", handler)

	const addr = ":8080"
	fmt.Printf("Start Listen %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, &ochttp.Handler{
		Handler:     mux,
		Propagation: &propagation.HTTPFormat{},
		FormatSpanName: func(req *http.Request) string {
			return fmt.Sprintf("/queue2%s", req.URL.Path)
		},
	}))
}

func createHandlers(ctx context.Context, projectID string, serviceAccountEmail string) (*Handlers, error) {
	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	helloTopicPubSubService, err := pubsub2.NewPubSubService(ctx, pubSubClient, "hello", projectID, false)
	if err != nil {
		return nil, err
	}
	helloOrderTopicPubSubService, err := pubsub2.NewPubSubService(ctx, pubSubClient, "hello-order", projectID, true)
	if err != nil {
		return nil, err
	}

	tasksClient, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	tasksboxService, err := cloudtasksbox.NewService(ctx, tasksClient, serviceAccountEmail)
	if err != nil {
		return nil, err
	}
	helloQueueTasksService, err := tasks2.NewTasksService(ctx, &cloudtasksbox.Queue{
		ProjectID: projectID,
		Region:    "asia-northeast1",
		Name:      "hello",
	}, tasksboxService)

	pubsubHelloHandler, err := pubsub2.NewHelloHandler(ctx, helloTopicPubSubService, helloOrderTopicPubSubService)
	if err != nil {
		return nil, err
	}
	pubsubReceiveHandler, err := pubsub2.NewReceiveHandler(ctx)
	if err != nil {
		return nil, err
	}

	tasksHelloHandler, err := tasks2.NewHelloHandler(ctx, helloQueueTasksService)
	if err != nil {
		return nil, err
	}
	tasksReceiveHandler, err := tasks2.NewReceiveHandler(ctx)
	if err != nil {
		return nil, err
	}

	return &Handlers{
		PubSubHelloHandler:   pubsubHelloHandler,
		PubSubReceiveHandler: pubsubReceiveHandler,
		TasksHelloHandler:    tasksHelloHandler,
		TasksReceiveHandler:  tasksReceiveHandler,
	}, nil
}
