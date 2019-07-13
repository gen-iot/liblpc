package liblpc

import (
	"gitee.com/Puietel/std"
	"syscall"
)

type ListenerOnAccept func(ln *Listener, newFd int, err error)

type Listener struct {
	*FdWatcher
	onAccept ListenerOnAccept
}

func NewListener(loop EventLoop, fd int, onAccept ListenerOnAccept) *Listener {
	std.Assert(onAccept != nil, "onAccept callback is nil")
	_ = syscall.SetNonblock(fd, true)
	l := new(Listener)
	l.FdWatcher = NewFdWatcher(loop, fd, l)
	l.onAccept = onAccept
	return l
}

func (this *Listener) OnEvent(event uint32) {
	if event&syscall.EPOLLIN == 0 {
		return
	}
	for {
		fd, _, err := syscall.Accept4(this.GetFd(), syscall.O_NONBLOCK|syscall.O_CLOEXEC)
		if err != nil {
			if WOULDBLOCK(err) {
				return
			}
			this.onAccept(this, -1, err)
			return
		}
		this.onAccept(this, fd, nil)
	}
}
