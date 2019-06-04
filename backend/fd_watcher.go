package backend

import (
	"syscall"
)

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

func (this *FdWatcher) Update(inLoop bool) {

}

func (this *FdWatcher) Close() error {
	return syscall.Close(this.fd)
}

func (this *FdWatcher) WantRead() (update bool) {
	if this.event&syscall.EPOLLIN != 0 {
		return false
	}
	this.event |= syscall.EPOLLIN
	return true
}

func (this *FdWatcher) DisableRead() (update bool) {
	if this.event&syscall.EPOLLIN != 0 {
		this.event &= (^uint32(syscall.EPOLLIN))
		return true
	}
	return false
}

func (this *FdWatcher) WantWrite() (update bool) {
	if this.event&syscall.EPOLLOUT != 0 {
		return false
	}
	this.event |= syscall.EPOLLOUT
	return true
}

func (this *FdWatcher) DisableWrite() (update bool) {
	if this.event&syscall.EPOLLOUT != 0 {
		this.event &= (^uint32(syscall.EPOLLOUT))
		return true
	}

	return false
}

func (this *FdWatcher) DisableRW() (update bool) {
	if this.event == 0 {
		return false
	}
	this.event = 0
	return true
}

func WOULDBLOCK(err error) bool {
	if err == nil {
		return false
	}
	return err == syscall.EAGAIN || err == syscall.EWOULDBLOCK
}

func (this *FdWatcher) OnEvent(event uint32) {
}
