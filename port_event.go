//+build !linux,!windows

package liblpc

import "golang.org/x/sys/unix"

const Readable = unix.POLLIN
const Writeable = unix.POLLOUT

type EventSizeType = int16
