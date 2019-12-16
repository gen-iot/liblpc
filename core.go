package liblpc

import (
	"io"
	"sync"
)

type EventWatcher interface {
	io.Closer
	GetFd() int
	GetEvent() EventSizeType
	SetEvent(event EventSizeType)
	Update(inLoop bool)
	OnEvent(event EventSizeType)
}

type PollerAction int

const (
	Add PollerAction = iota
	Mod
	Del
)

type Poller interface {
	io.Closer
	WatcherCtl(action PollerAction, watcher EventWatcher) error
	Poll(msec int) error
}

type FdWatcherMap struct {
	wm *sync.Map
}

func NewFdWatcherMap() *FdWatcherMap {
	return &FdWatcherMap{wm: new(sync.Map)}
}

func (this *FdWatcherMap) RmFd(fd int) {
	this.wm.Delete(fd)
}

func (this *FdWatcherMap) SetFd(fd int, watcher EventWatcher) {
	this.wm.Store(fd, watcher)
}

func (this *FdWatcherMap) GetWatcher(fd int) EventWatcher {
	value, ok := this.wm.Load(fd)
	if !ok {
		return nil
	}
	return value.(EventWatcher)
}

func (this *FdWatcherMap) Close() error {
	this.wm.Range(func(key, value interface{}) bool {
		_ = value.(EventWatcher).Close()
		return true
	})
	return nil
}
