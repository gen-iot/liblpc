package liblpc

import "gitee.com/Puietel/std"

type BufferedStream struct {
	*Stream
	bytesBuffer      *std.ByteBuffer
	onBufferedReadCb BufferedStreamOnRead
}

type BufferedStreamOnRead func(sw StreamWriter, buf std.ReadableBuffer, err error)

func NewBufferedConnStream(loop *IOEvtLoop, fd int, onRead BufferedStreamOnRead) *BufferedStream {
	s := new(BufferedStream)
	s.Stream = NewConnStream(loop, fd, s.onPlainStreamRead)
	s.SetWatcher(s)
	s.bytesBuffer = std.NewByteBuffer()
	s.onBufferedReadCb = onRead
	return s
}

func NewBufferedClientStream(loop *IOEvtLoop, fd int, onRead BufferedStreamOnRead) *BufferedStream {
	s := new(BufferedStream)
	s.Stream = NewClientStream(loop, fd, s.onPlainStreamRead)
	s.SetWatcher(s)
	s.bytesBuffer = std.NewByteBuffer()
	s.onBufferedReadCb = onRead
	return s
}

func (this *BufferedStream) onPlainStreamRead(sw StreamWriter, data []byte, len int, err error) {
	if err != nil {
		this.onBufferedReadCbWrapper(this, this.bytesBuffer, err)
		return
	}
	this.bytesBuffer.Write(data[:len])
	this.onBufferedReadCbWrapper(this, this.bytesBuffer, nil)
}

func (this *BufferedStream) onBufferedReadCbWrapper(sw StreamWriter, buf std.ReadableBuffer, err error) {
	if this.onBufferedReadCb == nil {
		return
	}
	this.onBufferedReadCb(sw, buf, err)
}
