// +build !linux,!windows

package liblpc

import (
	"github.com/gen-iot/std"
	"golang.org/x/sys/unix"
)

type Poll struct {
	fds      []unix.PollFd
	watchers *FdWatcherMap
}

func NewPoll() Poller {
	return &Poll{
		fds:      make([]unix.PollFd, 0, 32),
		watchers: NewFdWatcherMap(),
	}
}

func (this *Poll) Close() error {
	std.CloseIgnoreErr(this.watchers)
	panic("implement me")
}

func (this *Poll) handleReadyFd(pfd *unix.PollFd) {
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

func (this *Poll) WatcherCtl(action PollerAction, watcher EventWatcher) error {

}

func (this *Poll) Poll(msec int) error {
	nReady, err := unix.Poll(this.fds, -1)
	if err != nil {
		return err
	}
	if nReady == 0 {
		return nil
	}
	for idx := 0; idx < len(this.fds); idx++ {
		if nReady == 0 {
			break
		}
		pfd := &this.fds[idx]
		if pfd.Revents == 0 {
			continue
		}
		nReady--
		this.handleReadyFd(pfd)
	}
	return nil
}
