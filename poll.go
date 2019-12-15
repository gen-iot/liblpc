package liblpc

import "golang.org/x/sys/unix"

type Poll struct {
}

func NewPoll() Poller {

}

func (this *Poll) Close() error {
	panic("implement me")
}

func (this *Poll) WatcherCtl(action PollerAction, watcher EventWatcher) error {
	unix.PollFd{
		Fd:      watcher.GetFd(),
		Events:  watcher.GetEvent(),
		Revents: 0,
	}
}

func (this *Poll) Poll(msec int) error {

}
