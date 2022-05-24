package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

var errCount int64 = 1

func main() {
	client := http.DefaultClient

	qau := os.Getenv("QUEUE2_ATTACK_URL")
	fmt.Printf("QUEUE2_ATTACK_URL=%s\n", qau)
	if qau == "" {
		fmt.Println("$QUEUE2_ATTACK_URL is required")
	}

	for {
		start := time.Now()
		req, err := http.NewRequest(http.MethodGet, qau, nil)
		if err != nil {
			fmt.Printf("failed http.NewRequest url=%s,err=%s\n", qau, err)
			time.Sleep(getErrSleepDuration(errCount))
			continue
		}
		params := req.URL.Query()
		params.Add("order", uuid.New().String())
		params.Add("workTimeMillisecond", fmt.Sprintf("%d", rand.Int31n(3000)))
		if rand.Intn(30) < 2 {
			params.Add("forceFail", fmt.Sprintf("%d", rand.Intn(4)))
		}
		req.URL.RawQuery = params.Encode()

		res, err := client.Do(req)
		if err != nil {
			fmt.Printf("failed send request url=%s,err=%s\n", qau, err)
			time.Sleep(getErrSleepDuration(errCount))
			errCount++
			continue
		}
		workTime := time.Now().Sub(start)
		fmt.Printf("%s:%s:%s:%s\n", time.Now(), res.Status, workTime, req.URL.String())
		time.Sleep(time.Duration(rand.Int63n(5000)) * time.Millisecond)
	}
}

func getErrSleepDuration(errCount int64) time.Duration {
	return 100*time.Millisecond + time.Duration(errCount*errCount)*time.Second + time.Duration(rand.Int63n(errCount))*time.Second
}
