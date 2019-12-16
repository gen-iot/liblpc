// +build linux

package liblpc

import (
	"github.com/gen-iot/std"
	"golang.org/x/sys/unix"
	"os"
	"testing"
	"time"
)

func TestTimerFd(t *testing.T) {
	clock := ClockMonotonic
	stdLog("pid = ", os.Getpid())
	now, err := ClockGetTime(clock)
	std.AssertError(err, "ClockGetTime")
	itmspec := new(ITimerSpec)
	itmspec.ItInterval.Sec = 5
	itmspec.ItInterval.Nsec = 0
	itmspec.ItValue.Sec = now.Sec + 5
	itmspec.ItValue.Nsec = now.Nsec
	tmfd, err := TimerFdCreate(clock, TmFdCloexec|TmFdNonblock)
	if err != nil {
		panic(err)
	}
	stdLog("tmfd = ", tmfd)
	err = unix.SetNonblock(tmfd, false)
	std.AssertError(err, "SetNonblock")
	defer func() {
		_ = unix.Close(tmfd)
	}()
	err = TimerFdSetTime(tmfd, TmFdTimerAbstime, itmspec, nil)
	std.AssertError(err, "TimerFdSetTime")
	buf := make([]byte, 8)
	tmForRead := new(ITimerSpec)
	idx := 0
	for {
		nread, err := unix.Read(tmfd, buf)
		std.AssertError(err, "Read")
		if nread != 8 {
			panic("nread!=8")
		}
		stdLog("now is -> ", time.Now().String())
		err = TimerFdGetTime(tmfd, tmForRead)
		std.AssertError(err, "TimerFdGetTime")
		stdLog("get time from tmfd -> ", *tmForRead)
		timespec, err := ClockGetTime(clock)
		std.AssertError(err, "ClockGetTime")
		if idx%2 == 0 {
			idx = 1
			std.AssertError(TimerFdSetTime(tmfd, TmFdTimerAbstime, &ITimerSpec{
				ItInterval: unix.Timespec{
					Sec:  0,
					Nsec: 0,
				},
				ItValue: unix.Timespec{
					Sec:  timespec.Sec + 3,
					Nsec: timespec.Nsec,
				},
			}, nil), "err 1")
		} else {
			idx = 0
			std.AssertError(TimerFdSetTime(tmfd, TmFdTimerAbstime, &ITimerSpec{
				ItInterval: unix.Timespec{
					Sec:  0,
					Nsec: 0,
				},
				ItValue: unix.Timespec{
					Sec:  timespec.Sec + 1,
					Nsec: timespec.Nsec,
				},
			}, nil), "err 2")
		}
	}
}
