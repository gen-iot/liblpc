package liblpc

const MSG_NOSIGNAL = unix.MSG_NOSIGNAL

func init() {
	DefaultLoopNotifyCreator = NewNotifyWatcher
}
