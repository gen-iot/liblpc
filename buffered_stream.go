package liblpc

import "gitee.com/gen-iot/std"

type BufferedStream struct {
	*Stream
	bytesBuffer      *std.ByteBuffer
	onBufferedReadCb BufferedStreamOnRead
}

type BufferedStreamOnRead func(sw StreamWriter, buf std.ReadableBuffer)

func NewBufferedConnStream(loop *IOEvtLoop, fd int, onRead BufferedStreamOnRead) *BufferedStream {
	s := new(BufferedStream)
	s.Stream = NewConnStream(loop, fd, s.onStreamRead)
	s.SetWatcher(s)
	s.bytesBuffer = std.NewByteBuffer()
	s.onBufferedReadCb = onRead
	return s
}

func NewBufferedClientStream(loop *IOEvtLoop, fd int, onRead BufferedStreamOnRead) *BufferedStream {
	s := new(BufferedStream)
	s.Stream = NewClientStream(loop, fd, s.onStreamRead)
	s.SetWatcher(s)
	s.bytesBuffer = std.NewByteBuffer()
	s.onBufferedReadCb = onRead
	return s
}

func (this *BufferedStream) onStreamRead(sw StreamWriter, data []byte, len int) {
	this.bytesBuffer.Write(data[:len])
	if this.onBufferedReadCb == nil {
		return
	}
	this.onBufferedReadCb(sw, this.bytesBuffer)
}
