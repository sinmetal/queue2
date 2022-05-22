package main

import (
	"context"
	"fmt"

	"log"
	"net/http"

	"cloud.google.com/go/pubsub"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/sinmetal/queue2"
	pubsub2 "github.com/sinmetal/queue2/pubsub"
	metadatabox "github.com/sinmetalcraft/gcpbox/metadata"
	"github.com/vvakame/sdlog/aelog"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

type Handlers struct {
	PubSubService  *pubsub2.PubSubService
	HelloHandler   *pubsub2.HelloHandler
	ReceiveHandler *pubsub2.ReceiveHandler
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, err := fmt.Fprintf(w, "Hello, Ironlizard")
	if err != nil {
		aelog.Errorf(ctx, "err=%+\nv", err)
	}
}

func main() {
	ctx := context.Background()

	pID := queue2.ProjectID()
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

	handlers, err := createHandlers(ctx, pID)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/pubsub/hello", handlers.HelloHandler.Handle)
	mux.HandleFunc("/pubsub/receive", handlers.ReceiveHandler.Handle)
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

func createHandlers(ctx context.Context, projectID string) (*Handlers, error) {
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
	helloHandler, err := pubsub2.NewHelloHandler(ctx, helloTopicPubSubService, helloOrderTopicPubSubService)
	if err != nil {
		return nil, err
	}
	receiveHandler, err := pubsub2.NewReceiveHandler(ctx)
	if err != nil {
		return nil, err
	}
	return &Handlers{
		HelloHandler:   helloHandler,
		ReceiveHandler: receiveHandler,
	}, nil
}
