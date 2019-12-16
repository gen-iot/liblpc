// +build darwin

package liblpc

import (
	"github.com/gen-iot/std"
	"golang.org/x/sys/unix"
)

type Kqueue struct {
	kfd      int
	watchers *FdWatcherMap
	evtBuf   []unix.Kevent_t
}

func NewKqueue(pollSize int) (Poller, error) {
	kfd, err := unix.Kqueue()
	if err != nil {
		return nil, err
	}
	return &Kqueue{
		kfd:      kfd,
		watchers: NewFdWatcherMap(),
		evtBuf:   make([]unix.Kevent_t, pollSize),
	}, nil
}

func (this *Kqueue) Close() error {
	std.CloseIgnoreErr(this.watchers)
	return unix.Close(this.kfd)
}

func (this *Kqueue) handleReadyFd(pfd *unix.PollFd) {
	defer func() { pfd.Revents = 0 }()
	fd := pfd.Fd
	watcher := this.watchers.GetWatcher(int(fd))
	if watcher == nil {
		// disable any event...
		pfd.Events = 0
		stdLog("poll unknown fd = ", fd, ", watcher not found")
		return
	}
	watcher.OnEvent(pfd.Revents)
}

func (this *Kqueue) WatcherCtl(action PollerAction, watcher EventWatcher) error {
	switch action {
	case Add, Mod:
		return this.kqueueCtl(unix.EV_ADD, watcher)
	case Del:
		return this.kqueueCtl(unix.EV_DELETE|unix.EV_DISABLE|unix.EV_DISPATCH,
			watcher)
	}
	return nil
}

func (this *Kqueue) Poll(msec int) error {
	var tout *unix.Timespec = nil
	if msec <= 0 {
		tout = nil
	} else {
		tout = &unix.Timespec{
			Sec:  0,
			Nsec: int64(msec * 10e6),
		}
	}
	nEvent, err := unix.Kevent(this.kfd, nil, this.evtBuf, tout)
	if err != nil {
		return err
	}
	for idx := 0; idx < nEvent; idx++ {
		kEv := this.evtBuf[idx]
		fd := int(kEv.Ident)
		watcher := this.watchers.GetWatcher(fd)
		if watcher == nil {
			stdLog("kqueue unknown fd = ", fd, ", watcher not found")
			continue
		}
		switch kEv.Filter {
		case unix.EVFILT_WRITE:
			watcher.OnEvent(Writeable)
		default:
			watcher.OnEvent(Readable) // trigger read...
		}
	}
	return nil
}

func (this *Kqueue) kqueueCtl(flags uint16, watcher EventWatcher) error {
	filters := make([]int16, 0, 2)
	evt := watcher.GetEvent()
	if evt&Readable != 0 {
		filters = append(filters, unix.EVFILT_READ)
	}
	if evt&Writeable != 0 {
		filters = append(filters, unix.EVFILT_WRITE)
	}
	fd := watcher.GetFd()
	ke := unix.Kevent_t{
		Ident: uint64(fd),
		Flags: flags,
	}
	if len(filters) == 0 {
		if _, err := unix.Kevent(this.kfd,
			[]unix.Kevent_t{ke}, nil, nil); err != nil {
			return err
		}
		return nil
	}
	for _, f := range filters {
		ke.Filter = f
		this.watchers.SetFd(fd, watcher)
		if _, err := unix.Kevent(this.kfd,
			[]unix.Kevent_t{ke}, nil, nil); err != nil {
			return err
		}
	}
	return nil
}

func NewDefaultPoller(pollSize int) (Poller, error) {
	return NewKqueue(pollSize)
}
