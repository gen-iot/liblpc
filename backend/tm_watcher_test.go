package backend

import (
	"fmt"
	"liblpc"
	"testing"
	"time"
)

func TestTMWatcher(t *testing.T) {
	loop, err := NewEventLoop()
	PanicIfError(err)
	defer func() { PanicIfError(loop.Close()) }()
	countDown := 5
	watcher, err := liblpc.NewTimerWatcher(loop, liblpc.ClockMonotonic, func(ins *liblpc.Timer) {
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
