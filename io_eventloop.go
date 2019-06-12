package liblpc

type IOEvtLoop struct {
	EventLoop
	ioBuffer []byte
}

const DefaultIOEvtLoopBufferSize = 1024 * 4

func NewIOEvtLoop(ioBufferSize int) (*IOEvtLoop, error) {
	var err error = nil
	l := new(IOEvtLoop)
	rawL, err := NewEventLoop()
	if err != nil {
		return nil, err
	}
	l.EventLoop = rawL
	l.ioBuffer = make([]byte, ioBufferSize, ioBufferSize)
	return l, nil
}
