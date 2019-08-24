package liblpc

import (
	"golang.org/x/sys/unix"
)

func WOULDBLOCK(err error) bool {
	if err == nil {
		return false
	}
	return err == unix.EAGAIN || err == unix.EWOULDBLOCK
}
