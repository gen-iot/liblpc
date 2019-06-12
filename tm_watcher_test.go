package liblpc

import (
	"fmt"
	"gitee.com/SuzhenProjects/liblpc/utils"
	"testing"
	"time"
)

func TestTMWatcher(t *testing.T) {
	loop, err := NewEventLoop()
	utils.PanicIfError(err)
	defer func() { utils.PanicIfError(loop.Close()) }()
	countDown := 5
	watcher, err := NewTimerWatcher(loop, ClockMonotonic, func(ins *Timer) {
		fmt.Println("ontick ", time.Now().String())
		countDown--
		if countDown == 0 {
			utils.PanicIfError(ins.Stop())
			loop.Break()
			fmt.Println("timer stop")
		}
	})
	utils.PanicIfError(err)
	defer func() { utils.PanicIfError(watcher.Close()) }()
	watcher.Update(true)
	utils.PanicIfError(err)
	err = watcher.Start(1000, 1000)
	loop.Run()
}
