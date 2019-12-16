package liblpc

import (
	"github.com/gen-iot/std"
	"golang.org/x/sys/unix"
)

type ListenerOnAccept func(ln *Listener, newFd int, err error)

type Listener struct {
	*FdWatcher
	onAccept ListenerOnAccept
}

func NewListener(loop EventLoop, fd int, onAccept ListenerOnAccept) *Listener {
	std.Assert(onAccept != nil, "onAccept callback is nil")
	_ = unix.SetNonblock(fd, true)
	l := new(Listener)
	l.FdWatcher = NewFdWatcher(loop, fd, l)
	l.onAccept = onAccept
	return l
}

func (this *Listener) OnEvent(event EventSizeType) {
	if event&Readable == 0 {
		return
	}
	for {
		fd, _, err := SockFd(this.GetFd()).Accept(unix.O_NONBLOCK | unix.O_CLOEXEC)
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
