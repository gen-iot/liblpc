package liblpc

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

/**
 * conn[0], for server side use
 * conn[1], can pass to child process
 */
func grabChildFd() (local net.Conn, remote net.Conn, err error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, nil, err
	}
	// mark 0 closeExec
	syscall.CloseOnExec(fds[0])
	conns := make([]net.Conn, 2)
	for idx := 0; idx < 2; idx++ {
		_ = syscall.SetNonblock(fds[idx], true)
		file := os.NewFile(uintptr(fds[idx]), fmt.Sprintf("lpc-fd-%d", fds[idx]))
		conns[idx], err = net.FileConn(file)
		if err != nil {
			return nil, nil, err
		}
	}
	return conns[0], conns[1], nil
}
