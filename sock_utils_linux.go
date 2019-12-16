package liblpc

import (
	"golang.org/x/sys/unix"
	"syscall"
)

func (this SockFd) Accept(flags int) (nfd int, sa unix.Sockaddr, err error) {
	return unix.Accept4(int(this), flags)
}

// fd[0] for parent process
// fd[1] for child process
// nonblock : set socket nonblock
func MakeIpcSockpair(nonblock bool) (fds [2]int, err error) {
	syscall.ForkLock.Lock()
	defer syscall.ForkLock.Unlock()
	typ := unix.SOCK_STREAM
	if nonblock {
		typ |= unix.O_NONBLOCK
	}
	fds, err =
		unix.Socketpair(unix.AF_UNIX,
			unix.SOCK_STREAM|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return
	}
	return
}
