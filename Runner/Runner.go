package Runner

import (
	"errors"
	"os"
	"os/signal"
	"time"
)

//ErrTimeOut 工作逾時
var ErrTimeOut = errors.New("工作時程逾時")

//ErrInterrupt 工作打斷
var ErrInterrupt = errors.New("工作已被打斷")

//Runner ，可排程執行工作，而且可以控制
type Runner struct {
	tasks     []func(int)      //func array to Do
	complete  chan error       //通知任務完成
	timeout   <-chan time.Time //所有任務最終的TimeOut
	interrupt chan os.Signal   //控制強制中止的訊號

}

// New 建構子
func New(tm time.Duration) *Runner {
	return &Runner{
		complete:  make(chan error),
		timeout:   time.After(tm),
		interrupt: make(chan os.Signal, 1),
	}
}

//Add (public)將要執行的工作添加到Runner準備執行
func (r *Runner) Add(tasks ...func(int)) {
	r.tasks = append(r.tasks, tasks...)
}

//run (private)開始執行工作，除非中斷否則會執行到沒有工作
func (r *Runner) run() error {
	for id, task := range r.tasks {
		if r.isInterrupt() {
			return ErrInterrupt
		}
		task(id)
	}
	return nil
}

//isInterrupt (private)檢查是否收到中斷訊號 ex:Ctrl + C
func (r *Runner) isInterrupt() bool {
	select {
	case <-r.interrupt:
		signal.Stop(r.interrupt)
		return true
	default:
		return false
	}
}

//Start (public)開始執行工作，並監聽中斷事件
func (r *Runner) Start() error {
	//希望接收哪些系统信号
	signal.Notify(r.interrupt, os.Interrupt)

	go func() {
		r.complete <- r.run()
	}()

	select {
	case err := <-r.complete:
		return err
	case <-r.timeout:
		return ErrTimeOut
	}
}
