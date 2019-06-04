package backend

import "syscall"

type baseEventWatcher struct {
	fd    int
	event uint32
}

func (this *baseEventWatcher) GetFd() int {
	return this.fd
}

func (this *baseEventWatcher) GetEvent() uint32 {
	return this.event
}

func (this *baseEventWatcher) SetEvent(event uint32) {
	this.event = event
}

type FdWatcher struct {
	baseEventWatcher
}

func NewFdWatcher(fd int, closeExec bool) *FdWatcher {
	if closeExec {
		syscall.CloseOnExec(fd)
	}
	_ = syscall.SetNonblock(fd, true)
	w := new(FdWatcher)
	w.fd = fd
	w.event = syscall.EPOLLIN
	return w
}

func (this *FdWatcher) Close() error {
	return syscall.Close(this.fd)
}

func (this *FdWatcher) WantRead() {
	this.event |= syscall.EPOLLIN
}

func (this *FdWatcher) WantWrite() {
	this.event |= syscall.EPOLLOUT
}

func (this *FdWatcher) OnRead() error {
	return nil
}

func (this *FdWatcher) OnWrite() error {
	return nil
}

func (this *FdWatcher) OnEvent(event uint32) {
	if event&syscall.EPOLLIN != 0 {
		//read
		for {
			err := this.OnRead()
		}
	}

	if event&syscall.EPOLLOUT != 0 {
		//writable
		for {
			err := this.OnWrite()
		}
	}
}
