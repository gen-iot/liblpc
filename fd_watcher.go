package liblpc

import (
	"github.com/gen-iot/std"
	"golang.org/x/sys/unix"
	"log"
	"sync/atomic"
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
	loop          EventLoop
	fd            int
	event         uint32
	attachToLoop  bool
	closeFlag     int32
	drivenWatcher IOWatcher
	BaseUserData
}

func NewFdWatcher(loop EventLoop, fd int, watcher IOWatcher) *FdWatcher {
	w := new(FdWatcher)
	w.loop = loop
	w.fd = fd
	w.event = 0
	w.attachToLoop = false
	w.closeFlag = 0
	w.drivenWatcher = watcher
	return w
}

func (this *FdWatcher) Start() {
	this.loop.RunInLoop(func() {
		this.drivenWatcher.WantRead()
		this.drivenWatcher.WantWrite()
		this.drivenWatcher.Update(true)
	})
}

// helper for driven class
func (this *FdWatcher) SetWatcher(watcher IOWatcher) {
	this.drivenWatcher = watcher
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
		if !this.attachToLoop && this.drivenWatcher.GetEvent() == 0 {
			return
		}
		mode := Mod
		if this.attachToLoop && this.drivenWatcher.GetEvent() == 0 {
			mode = Del
			this.attachToLoop = false
		} else if !this.attachToLoop {
			mode = Add
			this.attachToLoop = true
		}
		err := this.loop.Poller().WatcherCtl(mode, this.drivenWatcher)
		if err != nil {
			log.Printf("fd_watcher watcherCtl fail(%v) watcher(memory:%v)\n", err, this.drivenWatcher)
			// close watcher
			std.CloseIgnoreErr(this.drivenWatcher)
		}
	} else {
		this.loop.RunInLoop(func() {
			this.drivenWatcher.Update(true)
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
		_ = this.loop.Poller().WatcherCtl(Del, this.drivenWatcher)
	}
	return unix.Close(this.fd)
}

func (this *FdWatcher) WantRead() (update bool) {
	if this.event&unix.EPOLLIN != 0 {
		return false
	}
	this.event |= unix.EPOLLIN
	return true
}

func (this *FdWatcher) DisableRead() (update bool) {
	if this.event&unix.EPOLLIN != 0 {
		this.event &= ^uint32(unix.EPOLLIN)
		return true
	}
	return false
}

func (this *FdWatcher) WantWrite() (update bool) {
	if this.event&unix.EPOLLOUT != 0 {
		return false
	}
	this.event |= unix.EPOLLOUT
	return true
}

func (this *FdWatcher) DisableWrite() (update bool) {
	if this.event&unix.EPOLLOUT != 0 {
		this.event &= ^uint32(unix.EPOLLOUT)
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
