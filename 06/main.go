package main

import (
	runner "Go-Learning/Runner"

	"fmt"
	"os"
	"sync"
	"time"

	greq "github.com/syhlion/greq"
	requestwork "github.com/syhlion/requestwork.v2"
)

func main() {

	fmt.Println("...工作開始...")

	timeout := 20 * time.Second
	r := runner.New(timeout)
	worker := requestwork.New(50)

	client := greq.New(worker, 15*time.Second, false)

	r.Add(DoWork(client), DoWork(client))

	if err := r.Start(); err != nil {
		switch err {
		case runner.ErrTimeOut:
			fmt.Println(err)
			os.Exit(1)
		case runner.ErrInterrupt:
			fmt.Println(err)
			os.Exit(2)
		}
	}
	fmt.Println("...工作結束...")
}

//stop := make(chan bool)

// DoWork 執行qreq測試
func DoWork(client *greq.Client) func(int) {
	return func(id int) {
		fmt.Printf("Mission%d\n", id)
		const (
			count = 10
		)
		var wg sync.WaitGroup
		wg.Add(count)

		for s := 0; s < count; s++ {
			go func(i int) {
				client.Get("https://tw.yahoo.com", nil)
				fmt.Println(i)
				wg.Done()
			}(s)
		}
		wg.Wait()

		// //POST
		// v := url.Values{}
		// v.Add("data", "123")
		// client.Post("https://tw.yahoo.com", v)

	}
}
