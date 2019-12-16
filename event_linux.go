// +build linux

package liblpc

import "golang.org/x/sys/unix"

const Readable = unix.EPOLLIN
const Writeable = unix.EPOLLOUT

type EventSizeType = uint32
