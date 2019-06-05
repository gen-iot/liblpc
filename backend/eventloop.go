package backend

import (
	"container/list"
	"io"
	"sync/atomic"
)

type EventLoop interface {
	io.Closer
	RunInLoop(cb func())
	Notify()
	Run()
	Break()
	Poller() Poller
}

type evtLoop struct {
	poller    Poller
	notify    *NotifyWatcher
	cbQ       *list.List
	lock      *SpinLock
	closeFlag int32
	stopFlag  int32
	endRunSig chan struct{}
}

func NewEventLoop() (EventLoop, error) {
	var err error = nil

	l := new(evtLoop)
	//
	l.poller, err = NewPoll()
	if err != nil {
		return nil, err
	}
	l.notify, err = NewNotifyWatcher(l, l.onWakeUp)
	if err != nil {
		return nil, err
	}
	//
	err = l.poller.WatcherCtl(Add, l.notify)
	if err != nil {
		_ = l.notify.Close()
		_ = l.poller.Close()
		return nil, err
	}
	l.cbQ = list.New()
	l.lock = new(SpinLock)
	l.closeFlag = 0
	l.stopFlag = 0
	l.endRunSig = make(chan struct{})
	return l, nil
}

func (this *evtLoop) RunInLoop(cb func()) {
	this.lock.Lock()
	if atomic.LoadInt32(&this.stopFlag) == 1 {
		this.lock.Unlock()
		cb()
		return
	}
	this.cbQ.PushBack(cb)
	this.lock.Unlock()
	this.Notify()
}

func (this *evtLoop) Notify() {
	this.notify.Notify()
}

func (this *evtLoop) onWakeUp() {
	this.processPending()
}

func (this *evtLoop) processPending() {
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

func (this *evtLoop) Break() {
	if atomic.LoadInt32(&this.stopFlag) == 1 {
		panic("loop already send stop signal!")
	}
	atomic.StoreInt32(&this.stopFlag, 1)
}

func (this *evtLoop) Close() error {
	<-this.endRunSig
	if atomic.CompareAndSwapInt32(&this.closeFlag, 0, 1) {
		_ = this.poller.Close()
		_ = this.notify.Close()
	}
	return nil
}

func (this *evtLoop) Poller() Poller {
	return this.poller
}

func (this *evtLoop) Run() {
	if atomic.LoadInt32(&this.stopFlag) == 1 {
		panic("loop already finished!")
	}
	for {
		_ = this.poller.Poll(-1)
		if atomic.LoadInt32(&this.stopFlag) == 1 {
			break
		}
	}
	this.processPending()
	close(this.endRunSig)
}
