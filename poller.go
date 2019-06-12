package liblpc

import "io"

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
