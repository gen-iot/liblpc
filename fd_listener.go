package liblpc

import (
	"gitee.com/SuzhenProjects/liblpc/utils"
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
	l.Loop().RunInLoop(func() {
		l.WantRead()
		l.Update(true)
	})
	return l
}

func (this *FdListener) OnEvent(event uint32) {
	if event&syscall.EPOLLIN == 0 || this.onAccept == nil {
		return
	}
	for {
		fd, _, err := syscall.Accept4(this.GetFd(), syscall.O_NONBLOCK|syscall.O_CLOEXEC) //review me!
		if err != nil {
			if utils.WOULDBLOCK(err) {
				return
			}
			return
		}
		this.onAccept(this, fd)
	}
}
