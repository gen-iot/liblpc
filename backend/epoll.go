//+build linux

package backend

import (
	"syscall"
)

type Epoll struct {
	pollFd   int
	watchers map[int]EventWatcher
	pollBuf  []syscall.EpollEvent
}

func NewPoll(pollSize int) (*Epoll, error) {
	epoFd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		_ = syscall.Close(epoFd)
		return nil, err
	}
	p := new(Epoll)
	p.pollFd = epoFd
	p.watchers = make(map[int]EventWatcher)
	p.pollBuf = make([]syscall.EpollEvent, pollSize)
	return p, nil
}

func (this *Epoll) Close() error {
	for _, w := range this.watchers {
		_ = w.Close()
	}
	return syscall.Close(this.pollFd)
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
	nevents, err := syscall.EpollWait(this.pollFd, this.pollBuf, msec)
	if err != nil && err != syscall.EINTR {
		return err
	}
	for idx := 0; idx < nevents; idx++ {
		epEvent := this.pollBuf[idx]
		fd := int(epEvent.Fd)
		watcher, ok := this.watchers[fd]
		if !ok {
			continue
		}
		watcher.OnEvent(epEvent.Events)
	}
	return nil
}

func (this *Epoll) AddFd(fd int, event uint32, watcher EventWatcher) error {
	epEvent := &syscall.EpollEvent{
		Events: event,
		Fd:     int32(fd),
	}
	err := syscall.EpollCtl(this.pollFd, syscall.EPOLL_CTL_ADD, fd, epEvent)
	if err != nil {
		return err
	}
	this.watchers[fd] = watcher
	return nil
}

func (this *Epoll) ModFd(fd int, event uint32) error {
	epEvent := &syscall.EpollEvent{
		Events: event,
		Fd:     int32(fd),
	}
	err := syscall.EpollCtl(this.pollFd, syscall.EPOLL_CTL_MOD, fd, epEvent)
	return err
}

func (this *Epoll) DelFd(fd int) error {
	err := syscall.EpollCtl(this.pollFd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		delete(this.watchers, fd)
	}
	return err
}
