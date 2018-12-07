package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync"

	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/syhlion/greq"
	requestwork "github.com/syhlion/requestwork.v2"
)

// type Demo struct {
// 	Data map[string]int
// 	Lock *sync.RWMutex
// }

// func (d Demo) Get(key string) int {
// 	d.Lock.Lock()
// 	defer d.Lock.Unlock()
// 	return d.Data[key]
// }

// func (d Demo) Set(key string, value int) {
// 	d.Lock.Lock()
// 	defer d.Lock.Unlock()
// 	d.Data[key] = value
// }

func main() {

	//c := make(map[string]int)
	// c := &Demo{
	// 	Data: make(map[string]int),
	// 	Lock: &sync.RWMutex{},
	// }
	// for i := 0; i < 100; i++ {
	// 	go func() {
	// 		for j := 0; j < 500000; j++ {
	// 			c.Set(fmt.Sprintf("%d", j), j)
	// 		}
	// 	}()
	// }
	// for len(c.Data) < 500001 {
	// 	fmt.Println(len(c.Data))

	// 	fmt.Println(c.Get("499999"))

	// 	time.Sleep(time.Second * 3)
	// }

	stop := make(chan bool)

	//need import https://github.com/syhlion/requestwork.v2
	worker := requestwork.New(50)

	client := greq.New(worker, 15*time.Second, false)

	// time.Sleep(time.Second * 60)

	var wg sync.WaitGroup

	//GET
	wg.Add(1)
	go func(stops <-chan bool) {
		defer wg.Done()
		consumer(client, stops)
		//print(" ", index+1)
	}(stop)
	waitForSignal()
	close(stop)
	fmt.Println("stopping all jobs!")
	wg.Wait()

	//POST
	v := url.Values{}
	v.Add("data", "123")
	client.Post("https://tw.yahoo.com", v)

}
func waitForSignal() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt)
	signal.Notify(sigs, syscall.SIGTERM)
	<-sigs
}

func consumer(client *greq.Client, stop <-chan bool) {

	for {
		select {
		case <-stop:
			fmt.Println("exit sub goroutine")
			return
		default:
			_, status, _ := client.Get("https://tw.yahoo.com", nil)
			//fmt.Println(status)
			log.WithFields(log.Fields{
				"status": status,
			}).Debug("A test")
			time.Sleep(500 * time.Millisecond)
			//fmt.Println(status + 123)
		}
	}
}
