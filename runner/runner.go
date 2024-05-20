package runner

import (
	"errors"
	"os"
	"os/signal"
	"time"
)

type Runner struct {
	interrupt chan os.Signal
	complete  chan error
	timeout   <-chan time.Time
	tasks     []func(int)
}

var (
	ErrTimeOut   = errors.New("received timeout")
	ErrInterrupt = errors.New("received interrupt")
)

func New(d time.Duration) *Runner {
	return &Runner{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		timeout:   time.After(d),
	}
}

func (r *Runner) Add(tasks ...func(int)) {
	r.tasks = append(r.tasks, tasks...)
}

func (r *Runner) Start() error {
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

// 当前用户
func (r *Runner) run() error {
	for id, task := range r.tasks {
		if r.gotInterrupt() {
			return ErrInterrupt
		}
		task(id)
	}
	return nil
}

// 当前进程收到了中断信号
func (r *Runner) gotInterrupt() bool {
	select {
	case <-r.interrupt:
		signal.Stop(r.interrupt)
		return true
	default:
		return false
	}
}
