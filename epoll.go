//+build linux

package liblpc

import (
	"fmt"
	"golang.org/x/sys/unix"
	"sync"
)

type Epoll struct {
	efd    int
	wm     *sync.Map
	evtbuf []unix.EpollEvent
}

func NewPoll(pollSize int) (*Epoll, error) {
	epoFd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		_ = unix.Close(epoFd)
		return nil, err
	}
	p := new(Epoll)
	p.efd = epoFd
	p.wm = new(sync.Map)
	p.evtbuf = make([]unix.EpollEvent, pollSize)
	return p, nil
}
func (this *Epoll) rmFd(fd int) {
	this.wm.Delete(fd)
}

func (this *Epoll) setFd(fd int, watcher EventWatcher) {
	this.wm.Store(fd, watcher)
}

func (this *Epoll) getWatcher(fd int) EventWatcher {
	value, ok := this.wm.Load(fd)
	if !ok {
		return nil
	}
	return value.(EventWatcher)
}

func (this *Epoll) Close() error {
	this.wm.Range(func(key, value interface{}) bool {
		_ = value.(EventWatcher).Close()
		return true
	})
	return unix.Close(this.efd)
}

func (this *Epoll) WatcherCtl(action PollerAction, watcher EventWatcher) error {
	switch action {
	case Add:
		return this.AddFd(watcher.GetFd(), watcher.GetEvent(), watcher)
	case Mod:
		return this.ModFd(watcher.GetFd(), watcher.GetEvent())
	case Del:
		return this.DelFd(watcher.GetFd())
	}
	return nil
}

func (this *Epoll) Poll(msec int) error {
	nevents, err := unix.EpollWait(this.efd, this.evtbuf, msec)
	if err != nil && err != unix.EINTR {
		return err
	}
	for idx := 0; idx < nevents; idx++ {
		epEvent := this.evtbuf[idx]
		fd := int(epEvent.Fd)
		watcher := this.getWatcher(fd)
		if watcher == nil {
			fmt.Println("unknown fd = ", fd, ", drivenWatcher not found")
			continue
		}
		watcher.OnEvent(epEvent.Events)
	}
	return nil
}

func (this *Epoll) AddFd(fd int, event uint32, watcher EventWatcher) error {
	epEvent := &unix.EpollEvent{
		Events: event,
		Fd:     int32(fd),
	}
	err := unix.EpollCtl(this.efd, unix.EPOLL_CTL_ADD, fd, epEvent)
	if err != nil {
		return err
	}
	this.setFd(fd, watcher)
	return nil
}

func (this *Epoll) ModFd(fd int, event uint32) error {
	epEvent := &unix.EpollEvent{
		Events: event,
		Fd:     int32(fd),
	}
	err := unix.EpollCtl(this.efd, unix.EPOLL_CTL_MOD, fd, epEvent)
	return err
}

func (this *Epoll) DelFd(fd int) error {
	err := unix.EpollCtl(this.efd, unix.EPOLL_CTL_DEL, fd, nil)
	if err == nil {
		this.rmFd(fd)
	}
	return err
}
