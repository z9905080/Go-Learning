package Runner

import (
	"fmt"
	"os"
	"testing"
	"time"
)

type Args struct {
	tasks []func(int)
}

func CreateDefaultTestData() *Args {

	var task []func(int)
	mTask := append(task, funS(), funS())
	//mArgs := args{tasks: mTask}
	return &Args{
		tasks: mTask,
	}
	// &Runner{
	// 	complete:  make(chan error),
	// 	timeout:   time.After(tm),
	// 	interrupt: make(chan os.Signal, 1),
	// }
}

func TestRunner_Add(t *testing.T) {

	argss := CreateDefaultTestData()
	tests := []struct {
		name string
		r    *Runner
		args *Args
	}{
		// TODO: Add test cases.
		{
			name: "1",
			r: &Runner{
				complete:  make(chan error),
				timeout:   time.After(time.Second * 20), //所有工作逾時時間
				interrupt: make(chan os.Signal, 1),
			},
			args: argss,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Add(tt.args.tasks...)
		})
	}
}

func addFunc() []func(int) {
	return []func(a int){}
}

func funS() func(int) {
	return func(a int) {
		fmt.Println("tt")
	}
}

func TestRunner_Start(t *testing.T) {
	tests := []struct {
		name    string
		r       *Runner
		wantErr bool
	}{
		{
			name:    "1",
			r:       &Runner{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Runner.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
