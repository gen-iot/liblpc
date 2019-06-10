package backend

import "syscall"

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func Assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}

func WOULDBLOCK(err error) bool {
	if err == nil {
		return false
	}
	return err == syscall.EAGAIN || err == syscall.EWOULDBLOCK
}

