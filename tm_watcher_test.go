// +build linux

package liblpc

import (
	"github.com/gen-iot/std"
	"testing"
	"time"
)

func TestTMWatcher(t *testing.T) {
	loop, err := NewEventLoop()
	std.AssertError(err, "NewEventLoop")
	defer func() {
		std.AssertError(loop.Close(), "close loop")
	}()
	countDown := 5
	watcher, err := NewTimerWatcher(loop, ClockMonotonic, func(ins *Timer) {
		stdLog("ontick ", time.Now().String())
		countDown--
		if countDown == 0 {
			std.AssertError(ins.Stop(), "stop drivenWatcher")
			loop.Break()
			stdLog("timer stop")
		}
	})
	std.AssertError(err, "new timer drivenWatcher")
	defer func() { std.AssertError(watcher.Close(), "drivenWatcher close") }()
	err = watcher.StartTimer(1000, 1000)
	std.AssertError(err, "drivenWatcher start")
	loop.Run()
}
