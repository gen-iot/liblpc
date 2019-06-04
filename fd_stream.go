package liblpc

import (
	"container/list"
	"io"
	"liblpc/backend"
	"syscall"
)

type StreamWriter interface {
	io.Closer
	Write(data []byte, inLoop bool)
}

type FdStream struct {
	*backend.FdWatcher
	loop         *IOEvtLoop
	writeQ       *list.List
	attachToLoop bool
	onReadCb     func(sw StreamWriter, data []byte, len int, err error)
}

func NewFdStream(loop *IOEvtLoop, fd int, closeExec bool, onRead func(sw StreamWriter, data []byte, len int, err error)) *FdStream {
	stream := new(FdStream)
	stream.FdWatcher = backend.NewFdWatcher(fd, closeExec)
	stream.loop = loop
	stream.writeQ = list.New()
	stream.attachToLoop = false
	stream.onReadCb = onRead
	return stream
}
func (this *FdStream) Close() error {
	this.loop.RunInLoop(func() {
		this.DisableRW()
		this.Update(true)
		_ = this.FdWatcher.Close()
	})
	return nil
}

func (this *FdStream) Write(data []byte, inLoop bool) {
	if inLoop {
		if this.writeQ.Len() == 0 {
			//write directly
			nWrite, err := syscall.SendmsgN(this.GetFd(), data, nil, nil, syscall.MSG_NOSIGNAL)
			if err != nil {
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
		this.loop.RunInLoop(func() {
			this.Write(data, true)
		})
	}
}

func (this *FdStream) OnRead(data []byte, len int, err error) {
	if this.onReadCb != nil {
		this.onReadCb(this, data, len, err)
	}
}

func (this *FdStream) Update(inLoop bool) {
	if inLoop {
		if !this.attachToLoop && this.GetEvent() == 0 {
			return
		}
		mode := backend.Mod
		if this.attachToLoop && this.GetEvent() == 0 {
			mode = backend.Del
			this.attachToLoop = false
		} else if !this.attachToLoop {
			mode = backend.Add
			this.attachToLoop = true
		}
		err := this.loop.poller.WatcherCtl(mode, this)
		if err != nil {
			panic(err)
		}
	} else {
		this.loop.RunInLoop(func() {
			this.Update(true)
		})
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
					if backend.WOULDBLOCK(err) {
						dataWillWrite = dataWillWrite[nWrite:]
						if this.WantWrite() {
							this.Update(true)
						}
						break
					}
					this.OnRead(nil, 0, err)
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
			nRead, _, err := syscall.Recvfrom(this.GetFd(), this.loop.ioBuffer, syscall.MSG_NOSIGNAL)
			if nRead == 0 {
				err = io.EOF
			}
			if err != nil {
				if backend.WOULDBLOCK(err) {
					if this.WantRead() {
						this.Update(true)
					}
					break
				}
				this.OnRead(nil, 0, err)
				if this.DisableRW() {
					this.Update(true)
				}
				return
			}
			this.OnRead(this.loop.ioBuffer, nRead, err)
		}
	}
}
