package backend

import "syscall"

type TimerWatcher struct {
}

const (
	kCLOCK_MONOTONIC = 0x01
)

func NewTimerWatcher() *TimerWatcher {
	r1, _, err := syscall.Syscall(syscall.SYS_TIMERFD_CREATE, kCLOCK_MONOTONIC, syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
}
