package liblpc

import (
	"golang.org/x/sys/unix"
	"syscall"
)

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
	for _, fd := range fds {
		// cmd.Exec will dup fd , so we must set all paired_fd to closeOnExec
		// duped fd doesnt contain O_CLOEXEC
		unix.CloseOnExec(fd)
		_ = unix.SetNonblock(fd, nonblock)
	}
	return
}
