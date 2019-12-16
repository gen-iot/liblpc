package liblpc

import (
	"golang.org/x/sys/unix"
	"syscall"
)

func (this SockFd) Accept(flags int) (nfd int, sa unix.Sockaddr, err error) {
	nfd, sa, err = unix.Accept(int(this))
	if err != nil {
		return
	} else {
		if flags&unix.O_NONBLOCK != 0 {
			if err = unix.SetNonblock(nfd, true); err != nil {
				_ = unix.Close(nfd)
				return 0, nil, err
			}
		}
		if flags&unix.O_CLOEXEC != 0 {
			unix.CloseOnExec(nfd)
		}
	}
	return
}

// fd[0] for parent process
// fd[1] for child process
// nonblock : set socket nonblock
func MakeIpcSockpair(nonblock bool) (fds [2]int, err error) {
	syscall.ForkLock.Lock()
	defer syscall.ForkLock.Unlock()
	fds, err = unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		return
	}
	for _, it := range fds {
		unix.CloseOnExec(it)
		if nonblock {
			if err := unix.SetNonblock(it, nonblock); err != nil {
				_ = unix.Close(fds[0])
				_ = unix.Close(fds[1])
				fds = [2]int{}
				return fds, err
			}
		}
	}
	return
}

func Socket(domain, typ, proto int) (fd int, err error) {
	fd, err = unix.Socket(domain, typ, proto)
	if err != nil {
		return
	}
	if err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_NOSIGPIPE, 1); err != nil {
		_ = unix.Close(fd)
		fd = -1
		return
	}
	return
}
