package main

import (
	"log"
	"os"
	"time"

	runner "userLibarary/Runner"
)

func main() {
	log.Println("...开始执行任务...")

	timeout := 20 * time.Second
	r := runner.New(timeout)

	r.Add(createTask(), createTask(), createTask(), createTask(), createTask(), createTask(), createTask())

	if err := r.Start(); err != nil {
		switch err {
		case runner.ErrTimeOut:
			log.Println(err)
			os.Exit(1)
		case runner.ErrInterrupt:
			log.Println(err)
			os.Exit(2)
		}
	}
	//time.Sleep(2 * time.Second)
	log.Println("...任务执行结束...")
}

func createTask() func(int) {
	return func(id int) {
		log.Printf("正在执行任务%d", id)
		time.Sleep(2 * time.Second)
	}
}
