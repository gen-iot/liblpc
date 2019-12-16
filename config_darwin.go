package liblpc

const MSG_NOSIGNAL = 0

func init() {
	DefaultLoopNotifyCreator = NewGenericLoopNotify
}
