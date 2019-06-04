package liblpc

import "syscall"

// fd[0] for parent process
// fd[1] for child process
// nonblock : set socket nonblock
func MakeIpcSockpair(nonblock bool) (fds [2]int, err error) {
	fds, err = syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return
	}
	for _, fd := range fds {
		// cmd.Exec will dup fd , so we must set all paired_fd to closeOnExec
		// duped fd doesnt contain O_CLOEXEC
		syscall.CloseOnExec(fd)
		_ = syscall.SetNonblock(fd, nonblock)
	}
	return
}
