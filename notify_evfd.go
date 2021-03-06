// +build linux

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

func NewNotifyWatcher(loop EventLoop, wakeupCb func()) (LoopNotify, error) {
	eventFd, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		return nil, err
	}
	watcher := new(NotifyWatcher)
	watcher.FdWatcher = NewFdWatcher(loop, eventFd, watcher)
	watcher.wakeupCb = wakeupCb
	watcher.readBuf = make([]byte, 8)
	return watcher, nil
}

func (this *NotifyWatcher) OnEvent(event EventSizeType) {
	_, err := unix.Read(this.GetFd(), this.readBuf)
	if err != nil {
		return
	}
	if this.wakeupCb != nil {
		this.wakeupCb()
	}
}

func (this *NotifyWatcher) GetEvent() EventSizeType {
	return Readable
}

func (this *NotifyWatcher) SetEvent(event EventSizeType) {
}

func (this *NotifyWatcher) Notify() {
	_, _ = unix.Write(this.GetFd(), __notifyWatcherSendBuf)
}
