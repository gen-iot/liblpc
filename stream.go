package liblpc

import (
	"container/list"
	"io"
	"syscall"
)

type StreamWriter interface {
	io.Closer
	Write(data []byte, inLoop bool)
	UserDataStorage
}

type StreamOnRead func(sw StreamWriter, data []byte, len int, err error)
type StreamOnConnect func(sw StreamWriter)

type StreamMode int

const (
	ModeConn StreamMode = iota
	ModeClient
)

type Stream struct {
	*FdWatcher
	writeQ     *list.List
	onReadCb   StreamOnRead
	onConnect  StreamOnConnect
	readBuffer []byte
	isClose    bool
	writeReady bool
	mode       StreamMode
}

func _____newFdStream(loop *IOEvtLoop,
	mode StreamMode, fd int,
	rcb StreamOnRead) *Stream {
	_ = syscall.SetNonblock(fd, true)
	stream := new(Stream)
	stream.FdWatcher = NewFdWatcher(loop, fd, stream)
	stream.readBuffer = loop.ioBuffer
	stream.writeQ = list.New()
	stream.mode = mode
	stream.onReadCb = rcb
	return stream
}

func NewConnStream(loop *IOEvtLoop, fd int, rcb StreamOnRead) *Stream {
	return _____newFdStream(loop, ModeConn, fd, rcb)
}

func NewClientStream(loop *IOEvtLoop, fd int, rcb StreamOnRead) *Stream {
	return _____newFdStream(loop, ModeClient, fd, rcb)
}

func (this *Stream) SetOnConnect(cb StreamOnConnect) {
	this.onConnect = cb
}

func (this *Stream) Close() error {
	this.Loop().RunInLoop(func() {
		this.DisableRW()
		this.Update(true)
		_ = this.FdWatcher.Close()
	})
	return nil
}

func (this *Stream) Write(data []byte, inLoop bool) {
	if inLoop {
		if this.isClose {
			//log.Println("Stream Write : closed , write will be drop")
			return
		}
		if this.writeQ.Len() == 0 && this.writeReady {
			//write directly
			nWrite, err := syscall.SendmsgN(this.GetFd(), data, nil, nil, syscall.MSG_NOSIGNAL)
			if err != nil {
				//log.Println("Stream Write , err is ->", err)
				return
			}
			//log.Println("Stream Write N ->", nWrite)
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

func (this *Stream) onRead(data []byte, len int, err error) {
	if err != nil {
		this.isClose = true
	}
	if this.onReadCb != nil {
		this.onReadCb(this, data, len, err)
	}
}

func (this *Stream) OnEvent(event uint32) {
	if event&syscall.EPOLLOUT != 0 {
		// invoke onConnect

		if !this.writeReady {

			// SEE:https://linux.die.net/man/2/connect
			// AT: Return Value:EINPROGRESS
			// DOC:The socket is nonblocking and the connection cannot be completed immediately.
			// It is possible to select(2) or poll(2) for completion by selecting the socket for writing.
			// After select(2) indicates writability,
			// use getsockopt(2) to read the SO_ERROR option at level SOL_SOCKET to determine
			// whether connect() completed successfully (SO_ERROR is zero)
			// or unsuccessfully (SO_ERROR is one of the usual error codes listed here, explaining the reason for the failure).

			if this.mode == ModeClient {
				soErr, err := syscall.GetsockoptInt(this.fd, syscall.SOL_SOCKET, syscall.SO_ERROR)
				var connectErr error = nil
				if err != nil {
					// getsockopt error
					connectErr = err
				} else if soErr != 0 {
					// socket conn error
					connectErr = syscall.Errno(soErr)
				}
				if connectErr != nil {
					this.onRead(nil, 0, connectErr)
					if this.DisableRW() {
						this.Update(true)
					}
					return
				}
			}

			this.writeReady = true
			if this.onConnect != nil {
				this.onConnect(this)
			}
		}

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
					if WOULDBLOCK(err) {
						//log.Println("Stream OnEvent SendmsgN WOULDBLOCK")
						dataWillWrite = dataWillWrite[nWrite:]
						front.Value = dataWillWrite
						if this.WantWrite() {
							this.Update(true)
						}
						break
					}
					//log.Println("Stream OnEvent SendmsgN got error ->", err)
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

				if WOULDBLOCK(err) {
					//log.Println("Stream OnEvent Recvfrom WOULDBLOCK")
					if this.WantRead() {
						this.Update(true)
					}
					break
				} else {
					//log.Println("Stream OnEvent Recvfrom error -> ", err)
				}
				this.onRead(nil, 0, err)
				if this.DisableRW() {
					this.Update(true)
				}
				return
			}
			if nRead == 0 {
				//log.Println("Stream OnEvent Recvfrom EOF")
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
