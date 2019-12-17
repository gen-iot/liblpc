package liblpc

import "golang.org/x/sys/unix"

const MSG_NOSIGNAL = unix.MSG_NOSIGNAL

func init() {
	DefaultLoopNotifyCreator = NewNotifyWatcher
}
