package liblpc

import (
	"syscall"
)

type ListenerOnAccept func(ln *Listener, newFd int)

type Listener struct {
	*FdWatcher
	onAccept ListenerOnAccept
}

func NewListener(loop EventLoop, fd int, onAccept ListenerOnAccept) *Listener {
	_ = syscall.SetNonblock(fd, true)
	l := new(Listener)
	l.FdWatcher = NewFdWatcher(loop, fd, l)
	l.onAccept = onAccept
	if l.onAccept == nil {
		return l
	}
	return l
}

func (this *Listener) OnEvent(event uint32) {
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
