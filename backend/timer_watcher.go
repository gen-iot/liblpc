package backend

import (
	"syscall"
	"unsafe"
)

type ClockId int

const (
	ClockRealtime  ClockId = 0x00
	ClockMonotonic ClockId = 0x01
)

const (
	TmFdNonblock     = syscall.O_NONBLOCK
	TmFdCloexec      = syscall.O_CLOEXEC
	TmFdTimerAbstime = 1 << 0
)

type TimerWatcher struct {
	*FdWatcher
	clockId ClockId
	readBuf []byte
	onTick  func(*TimerWatcher)
}

func NewTimerWatcher(loop EventLoop, clockId ClockId, onTick func(*TimerWatcher)) (*TimerWatcher, error) {
	tfd, err := TimerFdCreate(clockId, TmFdNonblock|TmFdCloexec)
	if err != nil {
		return nil, err
	}
	tw := new(TimerWatcher)
	tw.FdWatcher = NewFdWatcher(loop, int(tfd), tw)
	tw.clockId = clockId
	tw.readBuf = make([]byte, 8)
	tw.onTick = onTick
	tw.WantRead()
	return tw, nil
}

func (this *TimerWatcher) OnEvent(event uint32) {
	if event&syscall.EPOLLIN == 0 {
		return
	}
	_, err := syscall.Read(this.GetFd(), this.readBuf)
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

func (this *TimerWatcher) Stop() error {
	return TimerFdSetTime(this.GetFd(),
		TmFdTimerAbstime,
		&ITimerSpec{},
		nil)
}

func (this *TimerWatcher) Start(delayMs int, intervalMs int) error {
	now, err := ClockGetTime(this.clockId)
	if err != nil {
		return err
	}

	delaySec := now.Sec
	delayNs := now.Nsec

	if delayMs != 0 {
		delayNs = int64(delayMs)*1e6 + now.Nsec
		if delayNs > 1e9 {
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
		if intervalNs > 1e9 {
			intervalSec = int64(intervalNs / 1e9)
			intervalNs = intervalNs % 1e9
		}
	}

	return TimerFdSetTime(this.GetFd(),
		TmFdTimerAbstime,
		&ITimerSpec{
			ItInterval: syscall.Timespec{
				Sec:  intervalSec,
				Nsec: intervalNs,
			},
			ItValue: syscall.Timespec{
				Sec:  delaySec,
				Nsec: delayNs,
			},
		}, nil)
}

type ITimerSpec struct {
	ItInterval syscall.Timespec
	ItValue    syscall.Timespec
}

func ClockGetTime(clockId ClockId) (*syscall.Timespec, error) {
	now := new(syscall.Timespec)
	_, _, err := syscall.Syscall(syscall.SYS_CLOCK_GETTIME, uintptr(clockId), uintptr(unsafe.Pointer(now)), 0)
	if err != 0 {
		return nil, syscall.Errno(err)
	}
	return now, nil
}

func TimerFdCreate(clockId ClockId, flags int) (int, error) {
	tmFd, _, err := syscall.Syscall(syscall.SYS_TIMERFD_CREATE, uintptr(clockId), uintptr(flags), 0)
	if err != 0 {
		return -1, syscall.Errno(err)
	}
	return int(tmFd), nil
}

func TimerFdSetTime(fd int, flags int, new *ITimerSpec, old *ITimerSpec) error {
	_, _, err := syscall.Syscall6(syscall.SYS_TIMERFD_SETTIME,
		uintptr(fd), uintptr(flags), uintptr(unsafe.Pointer(new)), uintptr(unsafe.Pointer(old)), 0, 0)
	if err != 0 {
		return syscall.Errno(err)
	}
	return nil
}

func TimerFdGetTime(fd int, curr *ITimerSpec) error {
	_, _, err := syscall.Syscall(syscall.SYS_TIMERFD_GETTIME, uintptr(fd), uintptr(unsafe.Pointer(curr)), 0)
	if err != 0 {
		return syscall.Errno(err)
	}
	return nil
}
