package liblpc

import "syscall"

func WOULDBLOCK(err error) bool {
	if err == nil {
		return false
	}
	return err == syscall.EAGAIN || err == syscall.EWOULDBLOCK
}
