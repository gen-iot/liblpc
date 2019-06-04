package liblpc

import (
	"fmt"
	"liblpc/backend"
	"time"
)

type EvtLoop struct {
	poller backend.Poller
	notify *backend.NotifyWatcher
}

func NewEvtLoop() (*EvtLoop, error) {
	var err error = nil
	l := new(EvtLoop)
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
	return l, nil
}

func (this *EvtLoop) Notify() {
	this.notify.Notify()
}

func (this *EvtLoop) onWakeUp() {
	fmt.Println(time.Now().String())
}

func (this *EvtLoop) Run() {
	for {
		_ = this.poller.Poll(-1)
	}
}
