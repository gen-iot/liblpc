package liblpc

import (
	"fmt"
	"github.com/gen-iot/std"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestTimerFd(t *testing.T) {
	clock := ClockMonotonic
	fmt.Println("pid = ", os.Getpid())
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
	fmt.Println("tmfd = ", tmfd)
	err = syscall.SetNonblock(tmfd, false)
	std.AssertError(err, "SetNonblock")
	defer func() {
		_ = syscall.Close(tmfd)
	}()
	err = TimerFdSetTime(tmfd, TmFdTimerAbstime, itmspec, nil)
	std.AssertError(err, "TimerFdSetTime")
	buf := make([]byte, 8)
	tmForRead := new(ITimerSpec)
	idx := 0
	for {
		nread, err := syscall.Read(tmfd, buf)
		std.AssertError(err, "Read")
		if nread != 8 {
			panic("nread!=8")
		}
		fmt.Println("now is -> ", time.Now().String())
		err = TimerFdGetTime(tmfd, tmForRead)
		std.AssertError(err, "TimerFdGetTime")
		fmt.Println("get time from tmfd -> ", *tmForRead)
		timespec, err := ClockGetTime(clock)
		std.AssertError(err, "ClockGetTime")
		if idx%2 == 0 {
			idx = 1
			TimerFdSetTime(tmfd, TmFdTimerAbstime, &ITimerSpec{
				ItInterval: syscall.Timespec{
					Sec:  0,
					Nsec: 0,
				},
				ItValue: syscall.Timespec{
					Sec:  timespec.Sec + 3,
					Nsec: timespec.Nsec,
				},
			}, nil)
		} else {
			idx = 0
			TimerFdSetTime(tmfd, TmFdTimerAbstime, &ITimerSpec{
				ItInterval: syscall.Timespec{
					Sec:  0,
					Nsec: 0,
				},
				ItValue: syscall.Timespec{
					Sec:  timespec.Sec + 1,
					Nsec: timespec.Nsec,
				},
			}, nil)
		}
	}
}
