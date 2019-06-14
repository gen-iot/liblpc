package liblpc

import "gitee.com/Puietel/std"

type FdBufferedStream struct {
	*FdStream
	bytesBuffer      *std.ByteBuffer
	onBufferedReadCb FdBufferedStreamOnRead
}

type FdBufferedStreamOnRead func(sw StreamWriter, buf std.ReadableBuffer, err error)

func NewFdBufferedStream(loop *IOEvtLoop, fd int, onRead FdBufferedStreamOnRead) *FdBufferedStream {
	s := new(FdBufferedStream)
	s.FdStream = NewFdStream(loop, fd, s.onFdStreamRead)
	s.SetWatcher(s)
	s.bytesBuffer = std.NewByteBuffer()
	s.onBufferedReadCb = onRead
	return s
}

func (this *FdBufferedStream) onFdStreamRead(sw StreamWriter, data []byte, len int, err error) {
	if err != nil {
		this.onBufferedReadCbWrapper(this, this.bytesBuffer, err)
		return
	}
	this.bytesBuffer.Write(data)
	this.onBufferedReadCbWrapper(this, this.bytesBuffer, nil)
}

func (this *FdBufferedStream) onBufferedReadCbWrapper(sw StreamWriter, buf std.ReadableBuffer, err error) {
	if this.onBufferedReadCb == nil {
		return
	}
	this.onBufferedReadCb(sw, buf, err)
}
