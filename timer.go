package liblpc

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

type ClockId int

const (
	ClockRealtime  ClockId = unix.CLOCK_REALTIME
	ClockMonotonic ClockId = unix.CLOCK_MONOTONIC
)

const (
	TmFdNonblock     = unix.O_NONBLOCK
	TmFdCloexec      = unix.O_CLOEXEC
	TmFdTimerAbstime = 1 << 0
)

type TimerOnTick func(*Timer)

type Timer struct {
	*FdWatcher
	clockId ClockId
	readBuf []byte
	onTick  TimerOnTick
}

func NewTimerWatcher(loop EventLoop, clockId ClockId, onTick TimerOnTick) (*Timer, error) {
	tfd, err := TimerFdCreate(clockId, TmFdNonblock|TmFdCloexec)
	if err != nil {
		return nil, err
	}
	tw := new(Timer)
	tw.FdWatcher = NewFdWatcher(loop, int(tfd), tw)
	tw.clockId = clockId
	tw.readBuf = make([]byte, 8)
	tw.onTick = onTick
	tw.Loop().RunInLoop(func() {
		tw.WantRead()
		tw.Update(true)
	})
	return tw, nil
}

func (this *Timer) OnEvent(event uint32) {
	if event&Readable == 0 {
		return
	}
	_, err := unix.Read(this.GetFd(), this.readBuf)
	if err != nil {
		if WOULDBLOCK(err) {
			if this.WantRead() {
				this.Update(true)
			}
		}
		return
	}
	if this.onTick != nil {
		this.onTick(this)
	}
}

func (this *Timer) Stop() error {
	return TimerFdSetTime(this.GetFd(),
		TmFdTimerAbstime,
		&ITimerSpec{},
		nil)
}

func (this *Timer) StartTimer(delayMs int, intervalMs int) error {
	now, err := ClockGetTime(this.clockId)
	if err != nil {
		return err
	}

	delaySec := now.Sec
	delayNs := now.Nsec

	if delayMs != 0 {
		delayNs = int64(delayMs)*1e6 + now.Nsec
		if delayNs >= 1e9 {
			extSec := delayNs / 1e9
			delaySec += extSec
			delayNs = delayNs % 1e9
		}
	}

	var intervalSec int64 = 0
	var intervalNs int64 = 0
	if intervalMs != 0 {
		intervalSec = int64(0)
		intervalNs = int64(intervalMs * 1e6)
		if intervalNs >= 1e9 {
			intervalSec = int64(intervalNs / 1e9)
			intervalNs = intervalNs % 1e9
		}
	}

	return TimerFdSetTime(this.GetFd(),
		TmFdTimerAbstime,
		&ITimerSpec{
			ItInterval: unix.Timespec{
				Sec:  intervalSec,
				Nsec: intervalNs,
			},
			ItValue: unix.Timespec{
				Sec:  delaySec,
				Nsec: delayNs,
			},
		}, nil)
}

type ITimerSpec struct {
	ItInterval unix.Timespec
	ItValue    unix.Timespec
}

func ClockGetTime(clockId ClockId) (*unix.Timespec, error) {
	now := new(unix.Timespec)
	_, _, err := unix.Syscall(unix.SYS_CLOCK_GETTIME, uintptr(clockId), uintptr(unsafe.Pointer(now)), 0)
	if err != 0 {
		return nil, err
	}
	return now, nil
}

func TimerFdCreate(clockId ClockId, flags int) (int, error) {
	tmFd, _, err := unix.Syscall(unix.SYS_TIMERFD_CREATE, uintptr(clockId), uintptr(flags), 0)
	if err != 0 {
		return -1, err
	}
	return int(tmFd), nil
}

func TimerFdSetTime(fd int, flags int, new *ITimerSpec, old *ITimerSpec) error {
	_, _, err := unix.Syscall6(unix.SYS_TIMERFD_SETTIME,
		uintptr(fd), uintptr(flags), uintptr(unsafe.Pointer(new)), uintptr(unsafe.Pointer(old)), 0, 0)
	if err != 0 {
		return unix.Errno(err)
	}
	return nil
}

func TimerFdGetTime(fd int, curr *ITimerSpec) error {
	_, _, err := unix.Syscall(unix.SYS_TIMERFD_GETTIME, uintptr(fd), uintptr(unsafe.Pointer(curr)), 0)
	if err != 0 {
		return err
	}
	return nil
}
