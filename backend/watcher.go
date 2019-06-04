package backend

import "io"

type Watcher interface {
	io.Closer
	OnEvent(event uint32)
	GetFd() int
	GetEvent() uint32
	SetEvent(event uint32)
}
