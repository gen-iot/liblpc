package liblpc

import (
	"errors"
	"fmt"
	"syscall"
)

var __notifyWatcherSendBuf []byte

func init() {
	__notifyWatcherSendBuf = make([]byte, 8)
	__notifyWatcherSendBuf[7] = 0x01
}

type NotifyWatcher struct {
	*FdWatcher
	wakeupCb func()
	readBuf  []byte
}

func NewNotifyWatcher(loop EventLoop, wakeupCb func()) (*NotifyWatcher, error) {
	eventFd, _, errno := syscall.Syscall(syscall.SYS_EVENTFD2, 0, syscall.O_CLOEXEC|syscall.O_NONBLOCK, 0)
	if errno != 0 {
		_ = syscall.Close(int(eventFd))
		return nil, errors.New(fmt.Sprintf("event fd failed err = %d", errno))
	}
	watcher := new(NotifyWatcher)
	watcher.FdWatcher = NewFdWatcher(loop, int(eventFd), watcher)
	watcher.wakeupCb = wakeupCb
	watcher.readBuf = make([]byte, 8)
	return watcher, nil
}

func (this *NotifyWatcher) OnEvent(event uint32) {
	_, err := syscall.Read(this.GetFd(), this.readBuf)
	if err != nil {
		return
	}
	if this.wakeupCb != nil {
		this.wakeupCb()
	}
}

func (this *NotifyWatcher) GetEvent() uint32 {
	return syscall.EPOLLIN
}

func (this *NotifyWatcher) SetEvent(event uint32) {
}

func (this *NotifyWatcher) Notify() {
	_, _ = syscall.Write(this.GetFd(), __notifyWatcherSendBuf)
}
