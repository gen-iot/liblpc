// +build darwin,!disable_kqueue

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
	case Add:
		return this.kqueueEvtAdd(watcher)
	case Mod:
		return this.kqueueEvtMod(watcher)
	case Del:
		return this.kqueueEvtDel(watcher)
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

type kqEvtHelper EventSizeType

func (this kqEvtHelper) Flag(testBit EventSizeType) uint16 {
	if EventSizeType(this)&testBit == 0 {
		return unix.EV_DELETE
	}
	return unix.EV_ADD
}

func (this *Kqueue) kqueueEvtAdd(watcher EventWatcher) error {
	kesAdd := make([]unix.Kevent_t, 0, 2)
	fd := watcher.GetFd()
	evt := watcher.GetEvent()

	if evt&Readable != 0 {
		kesAdd = append(kesAdd, unix.Kevent_t{
			Ident:  uint64(fd),
			Filter: unix.EVFILT_READ,
			Flags:  unix.EV_ADD,
		})
	}
	if evt&Writeable != 0 {
		kesAdd = append(kesAdd, unix.Kevent_t{
			Ident:  uint64(fd),
			Filter: unix.EVFILT_WRITE,
			Flags:  unix.EV_ADD,
		})
	}
	this.watchers.SetFd(fd, watcher)
	if len(kesAdd) != 0 {
		if _, err := unix.Kevent(this.kfd, kesAdd, nil, nil); err != nil {
			return err
		}
	}
	return nil
}

func (this *Kqueue) kqueueEvtMod(watcher EventWatcher) error {
	fd := watcher.GetFd()
	evt := watcher.GetEvent()
	evtOut := make([]unix.Kevent_t, 2)
	nErr, _ := unix.Kevent(this.kfd, []unix.Kevent_t{
		{
			Ident:  uint64(fd),
			Filter: unix.EVFILT_READ,
			Flags:  kqEvtHelper(evt).Flag(Readable),
		},
		{
			Ident:  uint64(fd),
			Filter: unix.EVFILT_WRITE,
			Flags:  kqEvtHelper(evt).Flag(Writeable),
		},
	}, evtOut, nil)
	if nErr == 0 {
		return nil
	}
	cmErrs := make(std.CombinedErrors, 0, 2)
	for idx := 0; idx < nErr; idx++ {
		if evtOut[idx].Flags&unix.EV_ERROR != 0 &&
			!unix.Errno(evtOut[idx].Data).Is(unix.ENOENT) {
			cmErrs = append(cmErrs, unix.Errno(evtOut[idx].Data))
		}
	}
	if len(cmErrs) == 0 {
		return nil
	}
	return cmErrs
}

func (this *Kqueue) kqueueEvtDel(watcher EventWatcher) error {
	kes := make([]unix.Kevent_t, 0, 2)
	fd := watcher.GetFd()
	evtOut := make([]unix.Kevent_t, 0, 2)
	kes = append(kes, unix.Kevent_t{
		Ident:  uint64(fd),
		Filter: unix.EVFILT_READ,
		Flags:  unix.EV_DISABLE | unix.EV_DELETE,
	})

	kes = append(kes, unix.Kevent_t{
		Ident:  uint64(fd),
		Filter: unix.EVFILT_WRITE,
		Flags:  unix.EV_DISABLE | unix.EV_DELETE,
	})
	this.watchers.RmFd(fd)
	nErr, _ := unix.Kevent(this.kfd, kes, evtOut, nil)
	if nErr == 0 {
		return nil
	}
	cmErrs := make(std.CombinedErrors, 0, 2)
	for idx := 0; idx < nErr; idx++ {
		if evtOut[idx].Flags&unix.EV_ERROR != 0 &&
			!unix.Errno(evtOut[idx].Data).Is(unix.ENOENT) {
			cmErrs = append(cmErrs, unix.Errno(evtOut[idx].Data))
		}
	}
	if len(cmErrs) == 0 {
		return nil
	}
	return cmErrs
}

func init() {
	DefaultPollerCreator = NewKqueue
}
