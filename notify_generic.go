package liblpc

import (
	"github.com/gen-iot/std"
	"golang.org/x/sys/unix"
)

type GenericLoopNotify struct {
	*FdWatcher
	wakeupCb func()
	readBuf  []byte
	sendData []byte
	wfd      int
}

func NewGenericLoopNotify(loop EventLoop, wakeupCb func()) (LoopNotify, error) {
	pairSock, err := MakeIpcSockpair(true)
	if err != nil {
		return nil, err
	}
	watcher := new(GenericLoopNotify)
	watcher.FdWatcher = NewFdWatcher(loop, pairSock[0], watcher)
	watcher.wakeupCb = wakeupCb
	watcher.readBuf = make([]byte, 64)
	watcher.sendData = make([]byte, 4)
	watcher.wfd = pairSock[1]
	return watcher, nil
}

func (this *GenericLoopNotify) OnEvent(event EventSizeType) {
	_, err := unix.Read(this.GetFd(), this.readBuf)
	if err != nil {
		return
	}
	if this.wakeupCb != nil {
		this.wakeupCb()
	}
}

func (this *GenericLoopNotify) GetEvent() EventSizeType {
	return Readable
}

func (this *GenericLoopNotify) SetEvent(event EventSizeType) {
}

func (this *GenericLoopNotify) Notify() {
	_, _ = unix.Write(this.wfd, this.sendData)
}

func (this *GenericLoopNotify) Close() error {
	std.CloseIgnoreErr(this.FdWatcher)
	_ = unix.Close(this.wfd)
	return nil
}
