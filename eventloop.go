package liblpc

import (
	"container/list"
	"liblpc/backend"
)

type EvtLoop struct {
	poller backend.Poller
	notify *backend.NotifyWatcher
	cbQ    *list.List
	lock   *backend.SpinLock
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
	l.cbQ = list.New()
	l.lock = new(backend.SpinLock)
	return l, nil
}

func (this *EvtLoop) RunInLoop(cb func()) {
	this.lock.Lock()
	this.cbQ.PushBack(cb)
	this.lock.Unlock()
	this.Notify()
}

func (this *EvtLoop) Notify() {
	this.notify.Notify()
}

func (this *EvtLoop) onWakeUp() {
	this.processPending()
}

func (this *EvtLoop) processPending() {
	this.lock.Lock()
	ls := this.cbQ
	this.cbQ = list.New()
	this.lock.Unlock()
	for ; ls.Len() != 0; {
		front := ls.Front()
		val := front.Value.(func())
		ls.Remove(front)
		val()
	}
}

func (this *EvtLoop) Run() {
	for {
		_ = this.poller.Poll(-1)
	}
	this.processPending()
}
