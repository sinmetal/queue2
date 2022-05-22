package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var errCount int64 = 1
var orderCount int64 = 0

func main() {
	client := http.DefaultClient

	qau := os.Getenv("QUEUE2_ATTACK_URL")
	fmt.Printf("QUEUE2_ATTACK_URL=%s\n", qau)
	if qau == "" {
		fmt.Println("$QUEUE2_ATTACK_URL is required")
	}

	for {
		start := time.Now()
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?order=%d", qau, orderCount), nil)
		if err != nil {
			fmt.Printf("failed http.NewRequest url=%s,err=%s\n", qau, err)
			time.Sleep(getErrSleepDuration(errCount))
			continue
		}
		res, err := client.Do(req)
		if err != nil {
			fmt.Printf("failed send request url=%s,err=%s\n", qau, err)
			time.Sleep(getErrSleepDuration(errCount))
			continue
		}
		workTime := time.Now().Sub(start)
		fmt.Printf("%s:%s:%s\n", time.Now(), res.Status, workTime)
		time.Sleep(time.Duration(rand.Int63n(3000)) * time.Millisecond)
		orderCount++
	}
}

func getErrSleepDuration(errCount int64) time.Duration {
	return 100*time.Millisecond + time.Duration(errCount*errCount)*time.Second + time.Duration(rand.Int63n(errCount))*time.Second
}
