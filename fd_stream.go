package liblpc

import (
	"container/list"
	"io"
	"log"
	"syscall"
)

type StreamWriter interface {
	io.Closer
	Write(data []byte, inLoop bool)
	UserDataStorage
}

type FdStreamOnRead func(sw StreamWriter, data []byte, len int, err error)

type FdStream struct {
	*FdWatcher
	writeQ     *list.List
	onReadCb   FdStreamOnRead
	readBuffer []byte
	isClose    bool
}

func NewFdStream(loop *IOEvtLoop, fd int, onRead FdStreamOnRead) *FdStream {
	_ = syscall.SetNonblock(fd, true)
	stream := new(FdStream)
	stream.FdWatcher = NewFdWatcher(loop, fd, stream)
	stream.readBuffer = loop.ioBuffer
	stream.writeQ = list.New()
	stream.onReadCb = onRead
	return stream
}

func (this *FdStream) Close() error {
	this.Loop().RunInLoop(func() {
		this.DisableRW()
		this.Update(true)
		_ = this.FdWatcher.Close()
	})
	return nil
}

func (this *FdStream) Write(data []byte, inLoop bool) {
	if inLoop {
		if this.isClose {
			log.Println("FdStream Write : closed , write will be drop")
			return
		}
		if this.writeQ.Len() == 0 {
			//write directly
			nWrite, err := syscall.SendmsgN(this.GetFd(), data, nil, nil, syscall.MSG_NOSIGNAL)
			if err != nil {
				log.Println("FdStream Write , err is ->", err)
				return
			}
			if nWrite != len(data) {
				data = data[nWrite:]
				this.writeQ.PushBack(data)
				if this.WantWrite() {
					this.Update(true)
				}
			}
		} else {
			this.writeQ.PushBack(data)
			if this.WantWrite() {
				this.Update(true)
			}
		}
	} else {
		this.Loop().RunInLoop(func() {
			this.Write(data, true)
		})
	}
}

func (this *FdStream) onRead(data []byte, len int, err error) {
	if err != nil {
		this.isClose = true
	}
	if this.onReadCb != nil {
		this.onReadCb(this, data, len, err)
	}
}

func (this *FdStream) OnEvent(event uint32) {
	if event&syscall.EPOLLOUT != 0 {
		if this.writeQ.Len() == 0 {
			if this.DisableWrite() {
				this.Update(true)
			}
		} else {
			//writable
			for ; this.writeQ.Len() != 0; {
				front := this.writeQ.Front()
				dataWillWrite := front.Value.([]byte)

				nWrite, err := syscall.SendmsgN(this.GetFd(), dataWillWrite, nil, nil, syscall.MSG_NOSIGNAL)
				if err != nil {
					log.Println("FdStream OnEvent SendmsgN got error ->", err)
					if WOULDBLOCK(err) {
						dataWillWrite = dataWillWrite[nWrite:]
						front.Value = dataWillWrite
						if this.WantWrite() {
							this.Update(true)
						}
						break
					}
					this.onRead(nil, 0, err)
					if this.DisableRW() {
						this.Update(true)
					}
					return
				}
				this.writeQ.Remove(front)
			}
		}
	}

	if event&syscall.EPOLLIN != 0 {
		//read
		for {
			nRead, _, err := syscall.Recvfrom(this.GetFd(), this.readBuffer, syscall.MSG_NOSIGNAL)
			if err != nil {
				log.Println("FdStream OnEvent Recvfrom error -> ", err)
				if WOULDBLOCK(err) {
					if this.WantRead() {
						this.Update(true)
					}
					break
				}
				this.onRead(nil, 0, err)
				if this.DisableRW() {
					this.Update(true)
				}
				return
			}
			if nRead == 0 {
				log.Println("FdStream OnEvent Recvfrom EOF")
				err = io.EOF
				this.onRead(nil, 0, err)
				if this.DisableRW() {
					this.Update(true)
				}
				return
			}
			this.onRead(this.readBuffer, nRead, err)
		}
	}
}
