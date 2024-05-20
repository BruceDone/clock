package runner

import (
	"testing"
	"time"
)

func TestRunner_Start(t *testing.T) {
	r := New(time.Duration(5) * time.Second)

	r.Add(func(i int) {
		t.Log("now will sleep 6 seconds")
		time.Sleep(time.Duration(1) * time.Second)
	})

	if err := r.Start(); err != nil {
		switch err {
		case ErrTimeOut:
			t.Log("time out")
		case ErrInterrupt:
			t.Log("signal got ")
		}
	}

	t.Log("process ended")
}

func TestSelect(t *testing.T) {
	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)

	select {
	case e1 := <-ch1:
		t.Logf("got 1 case1 %v", e1)
	case e2 := <-ch2:
		t.Logf("got 2 case1 %v", e2)
	}
}
