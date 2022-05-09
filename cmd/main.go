package main

import (
	"fmt"
	"log"
	"net/http"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	metadatabox "github.com/sinmetalcraft/gcpbox/metadata"
	"github.com/vvakame/sdlog/aelog"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, err := fmt.Fprintf(w, "Hello, Ironlizard")
	if err != nil {
		aelog.Errorf(ctx, "err=%+\nv", err)
	}
}

func main() {
	// ctx := context.Background()

	pID, err := metadatabox.ProjectID()
	if err != nil {
		panic(err)
	}
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

	mux := http.NewServeMux()
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
