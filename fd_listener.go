package liblpc

import (
	"syscall"
)

type FdListenerOnAccept func(ln *FdListener, newFd int)

type FdListener struct {
	*FdWatcher
	onAccept FdListenerOnAccept
}

func NewFdListener(loop EventLoop, fd int, onAccept FdListenerOnAccept) *FdListener {
	_ = syscall.SetNonblock(fd, true)
	l := new(FdListener)
	l.FdWatcher = NewFdWatcher(loop, fd, l)
	l.onAccept = onAccept
	if l.onAccept == nil {
		return l
	}
	return l
}

func (this *FdListener) OnEvent(event uint32) {
	if event&syscall.EPOLLIN == 0 || this.onAccept == nil {
		return
	}
	for {
		fd, _, err := syscall.Accept4(this.GetFd(), syscall.O_NONBLOCK|syscall.O_CLOEXEC)
		if err != nil {
			if WOULDBLOCK(err) {
				return
			}
			return
		}
		this.onAccept(this, fd)
	}
}
