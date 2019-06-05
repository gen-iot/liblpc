package backend

import (
	"fmt"
	"testing"
	"time"
)

func TestTMWatcher(t *testing.T) {
	loop, err := NewEventLoop()
	PanicIfError(err)
	countDown := 5
	watcher, err := NewTimerWatcher(loop, ClockMonotonic, func(ins *TimerWatcher) {
		fmt.Println("ontick ", time.Now().String())
		countDown--
		if countDown == 0 {
			PanicIfError(ins.Stop())
			fmt.Println("timer stop")
		}
	})
	defer func() {
		_ = watcher.Close()
	}()
	PanicIfError(err)
	watcher.Update(true)
	PanicIfError(err)
	err = watcher.Start(5000, 5000)
	loop.Run()
}
