package liblpc

import (
	"sync/atomic"
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
	closeFlag    int32
	watcher      IOWatcher
	BaseUserData
}

func NewFdWatcher(loop EventLoop, fd int, watcher IOWatcher) *FdWatcher {
	w := new(FdWatcher)
	w.loop = loop
	w.fd = fd
	w.event = 0
	w.attachToLoop = false
	w.closeFlag = 0
	w.watcher = watcher
	return w
}

func (this *FdWatcher) Start() {
	this.Loop().RunInLoop(func() {
		this.WantRead()
		this.Update(true)
	})
}

func (this *FdWatcher) StartConnect() {
	this.Loop().RunInLoop(func() {
		this.WantRead()
		this.WantWrite()
		this.Update(true)
	})
}

// helper for driven class
func (this *FdWatcher) SetWatcher(watcher IOWatcher) {
	this.watcher = watcher
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
		err := this.Loop().Poller().WatcherCtl(mode, this.watcher)
		if err != nil {
			panic(err)
		}
	} else {
		this.Loop().RunInLoop(func() {
			this.Update(true)
		})
	}
}

func (this *FdWatcher) Loop() EventLoop {
	return this.loop
}

func (this *FdWatcher) Close() error {
	if !atomic.CompareAndSwapInt32(&this.closeFlag, 0, 1) {
		return nil
	}
	if this.attachToLoop {
		_ = this.Loop().Poller().WatcherCtl(Del, this.watcher)
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
		this.event &= ^uint32(syscall.EPOLLIN)
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
		this.event &= ^uint32(syscall.EPOLLOUT)
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
