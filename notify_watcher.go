package liblpc

import (
	"golang.org/x/sys/unix"
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
	eventFd, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_CLOEXEC)
	if err != nil {
		return nil, err
	}
	watcher := new(NotifyWatcher)
	watcher.FdWatcher = NewFdWatcher(loop, int(eventFd), watcher)
	watcher.wakeupCb = wakeupCb
	watcher.readBuf = make([]byte, 8)
	return watcher, nil
}

func (this *NotifyWatcher) OnEvent(event uint32) {
	_, err := unix.Read(this.GetFd(), this.readBuf)
	if err != nil {
		return
	}
	if this.wakeupCb != nil {
		this.wakeupCb()
	}
}

func (this *NotifyWatcher) GetEvent() uint32 {
	return unix.EPOLLIN
}

func (this *NotifyWatcher) SetEvent(event uint32) {
}

func (this *NotifyWatcher) Notify() {
	_, _ = unix.Write(this.GetFd(), __notifyWatcherSendBuf)
}
