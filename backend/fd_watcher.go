package backend

import (
	"container/list"
	"syscall"
)

type baseEventWatcher struct {
	fd    int
	event uint32
}

func (this *baseEventWatcher) GetFd() int {
	return this.fd
}

func (this *baseEventWatcher) GetEvent() uint32 {
	return this.event
}

func (this *baseEventWatcher) SetEvent(event uint32) {
	this.event = event
}

type FdWatcher struct {
	baseEventWatcher
	readBuffer []byte
	writeQ     *list.List
}

func NewFdWatcher(fd int, closeExec bool, readBuffer []byte) *FdWatcher {
	if closeExec {
		syscall.CloseOnExec(fd)
	}
	_ = syscall.SetNonblock(fd, true)
	w := new(FdWatcher)
	w.fd = fd
	w.event = syscall.EPOLLIN
	if len(readBuffer) == 0 {
		readBuffer = make([]byte, 1024*4)
	}
	w.readBuffer = readBuffer
	w.writeQ = list.New()
	return w
}

func (this *FdWatcher) Close() error {
	return syscall.Close(this.fd)
}

func (this *FdWatcher) WantRead() {
	this.event |= syscall.EPOLLIN
}

func (this *FdWatcher) WantWrite() {
	this.event |= syscall.EPOLLOUT
}

func (this *FdWatcher) OnRead(data []byte, len int) error {
	return nil
}

func (this *FdWatcher) Write(data []byte) error {
	this.writeQ.PushBack(data)
	return nil
}

func WOULDBLOCK(err error) bool {
	if err == nil {
		return false
	}
	return err == syscall.EWOULDBLOCK || err == syscall.EAGAIN
}

func (this *FdWatcher) OnEvent(event uint32) {

	if event&syscall.EPOLLOUT != 0 {
		//writable
		for ; this.writeQ.Len() != 0; {
			front := this.writeQ.Front()
			dataWillWrite := front.Value.([]byte)

			nWrite, err := syscall.SendmsgN(this.fd, dataWillWrite, nil, nil, syscall.MSG_NOSIGNAL)
			if err != nil {
				if WOULDBLOCK(err) {
					dataWillWrite = dataWillWrite[nWrite:]
					break
				}
				//todo onclose
				return
			}

			this.writeQ.Remove(front)
		}
	}

	if event&syscall.EPOLLIN != 0 {
		//read
		for {
			nRead, _, err := syscall.Recvfrom(this.fd, this.readBuffer, syscall.MSG_NOSIGNAL)
			if err != nil && WOULDBLOCK(err) {
				//todo onclose
				return
			}

			err = this.OnRead(this.readBuffer, nRead)
			if err != nil {
				return
			}

		}
	}

}
