package backend

import (
	"syscall"
)

type IOWatcher interface {
	EventWatcher
	Loop() EventLoop
	WantRead() (update bool)
	DisableRead() (update bool)
	WantWrite() (update bool)
	DisableWrite() (update bool)
	DisableRW() (update bool)
}

type FdWatcher struct {
	loop         EventLoop
	fd           int
	event        uint32
	attachToLoop bool
	impl         IOWatcher
}

func NewFdWatcher(loop EventLoop, fd int, impl IOWatcher) *FdWatcher {
	w := new(FdWatcher)
	w.loop = loop
	w.fd = fd
	w.event = 0
	w.attachToLoop = false
	w.impl = impl
	return w
}

func (this *FdWatcher) GetFd() int {
	return this.fd
}

func (this *FdWatcher) GetEvent() uint32 {
	return this.event
}

func (this *FdWatcher) SetEvent(event uint32) {
	this.event = event
}

func (this *FdWatcher) Update(inLoop bool) {
	if inLoop {
		if !this.attachToLoop && this.GetEvent() == 0 {
			return
		}
		mode := Mod
		if this.attachToLoop && this.GetEvent() == 0 {
			mode = Del
			this.attachToLoop = false
		} else if !this.attachToLoop {
			mode = Add
			this.attachToLoop = true
		}
		err := this.loop.Poller().WatcherCtl(mode, this.impl)
		if err != nil {
			panic(err)
		}
	} else {
		this.loop.RunInLoop(func() {
			this.Update(true)
		})
	}
}

func (this *FdWatcher) Loop() EventLoop {
	return this.loop
}

func (this *FdWatcher) Close() error {
	if this.attachToLoop {
		_ = this.Loop().Poller().WatcherCtl(Del, this.impl)
	}
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
