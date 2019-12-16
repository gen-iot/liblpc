package liblpc

import (
	"github.com/gen-iot/std"
	"golang.org/x/sys/unix"
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
	event         EventSizeType
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

func (this *FdWatcher) GetEvent() EventSizeType {
	return this.event
}

func (this *FdWatcher) SetEvent(event EventSizeType) {
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
			stdLogf("fd_watcher watcherCtl fd(%d) fail(%v) watcher(memory:%v)\n", this.fd, err, this.drivenWatcher)
			// mark detachToLoop
			this.attachToLoop = false
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
	// stdLog("loop close fd ->", this.fd)
	return unix.Close(this.fd)
}

func (this *FdWatcher) WantRead() (update bool) {
	if this.event&Readable != 0 {
		return false
	}
	this.event |= Readable
	return true
}

func (this *FdWatcher) DisableRead() (update bool) {
	if this.event&Readable != 0 {
		this.event &= ^EventSizeType(Readable)
		return true
	}
	return false
}

func (this *FdWatcher) WantWrite() (update bool) {
	if this.event&Writeable != 0 {
		return false
	}
	this.event |= Writeable
	return true
}

func (this *FdWatcher) DisableWrite() (update bool) {
	if this.event&Writeable != 0 {
		this.event &= ^EventSizeType(Writeable)
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
