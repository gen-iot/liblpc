package backend

import (
	"fmt"
	"testing"
	"time"
)

func TestTMWatcher(t *testing.T) {
	loop, err := NewEventLoop()
	PanicIfError(err)
	defer func() { PanicIfError(loop.Close()) }()
	countDown := 5
	watcher, err := NewTimerWatcher(loop, ClockMonotonic, func(ins *TimerWatcher) {
		fmt.Println("ontick ", time.Now().String())
		countDown--
		if countDown == 0 {
			PanicIfError(ins.Stop())
			loop.Break()
			fmt.Println("timer stop")
		}
	})
	PanicIfError(err)
	defer func() { PanicIfError(watcher.Close()) }()
	watcher.Update(true)
	PanicIfError(err)
	err = watcher.Start(1000, 1000)
	loop.Run()
}
