//+build linux

package liblpc

import (
	"github.com/gen-iot/std"
	"golang.org/x/sys/unix"
)

type Epoll struct {
	efd      int
	watchers *FdWatcherMap
	evtbuf   []unix.EpollEvent
}

func NewEpoll(pollSize int) (Poller, error) {
	epoFd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		_ = unix.Close(epoFd)
		return nil, err
	}
	p := new(Epoll)
	p.efd = epoFd
	p.watchers = NewFdWatcherMap()
	p.evtbuf = make([]unix.EpollEvent, pollSize)
	return p, nil
}

func (this *Epoll) Close() error {
	std.CloseIgnoreErr(this.watchers)
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
		watcher := this.watchers.GetWatcher(fd)
		if watcher == nil {
			stdLog("epoll unknown fd = ", fd, ", watcher not found")
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
	this.watchers.SetFd(fd, watcher)
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
		this.watchers.RmFd(fd)
	}
	return err
}
