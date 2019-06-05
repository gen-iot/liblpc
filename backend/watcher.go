package backend

import "io"

type EventWatcher interface {
	io.Closer
	GetFd() int
	GetEvent() uint32
	SetEvent(event uint32)
	Update(inLoop bool)
	OnEvent(event uint32)
}
