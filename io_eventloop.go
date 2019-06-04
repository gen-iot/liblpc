package liblpc

import (
	"container/list"
	"liblpc/backend"
)

type IOEvtLoop struct {
	EvtLoop
	ioBuffer []byte
}

func NewIOEvtLoop(ioBufferSize int) (*IOEvtLoop, error) {
	var err error = nil

	l := new(IOEvtLoop)
	//
	l.poller, err = backend.NewPoll()
	if err != nil {
		return nil, err
	}
	l.notify, err = backend.NewNotifyWatcher(l.onWakeUp)
	if err != nil {
		return nil, err
	}
	//
	err = l.poller.WatcherCtl(backend.Add, l.notify)
	if err != nil {
		return nil, err
	}
	l.cbQ = list.New()
	l.lock = new(backend.SpinLock)
	l.ioBuffer = make([]byte, ioBufferSize)
	return l, nil
}
