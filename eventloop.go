package liblpc

import (
	"container/list"
	"context"
	"github.com/gen-iot/std"
	"io"
	"sync/atomic"
)

type EventLoop interface {
	io.Closer
	RunInLoop(cb func())
	Notify()
	Run(ctx context.Context)
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
	poller, err := NewEpoll(1024)
	if err != nil {
		return nil, err
	}
	return NewEventLoop2(poller)
}

func NewEventLoop2(poller Poller) (EventLoop, error) {
	var err error = nil

	l := new(evtLoop)
	//
	l.poller = poller
	//
	l.notify, err = NewNotifyWatcher(l, l.onWakeUp)
	if err != nil {
		std.CloseIgnoreErr(l.poller)
		return nil, err
	}
	l.notify.Update(true)
	//
	l.cbQ = list.New()
	l.lock = NewSpinLock()
	l.closeFlag = 0
	l.stopFlag = 0
	l.endRunSig = make(chan struct{})
	return l, nil
}

func (this *evtLoop) RunInLoop(cb func()) {
	if atomic.LoadInt32(&this.stopFlag) == 1 {
		cb()
		return
	}
	this.lock.Lock()
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
	for ls.Len() != 0 {
		front := ls.Front()
		val := front.Value.(func())
		ls.Remove(front)
		val()
	}
}

func (this *evtLoop) Break() {
	if atomic.LoadInt32(&this.stopFlag) == 1 {
		stdLog("note: loop already send stop signal.")
		return
	}
	atomic.StoreInt32(&this.stopFlag, 1)
	this.Notify()
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

func (this *evtLoop) Run(ctx context.Context) {
	if atomic.LoadInt32(&this.stopFlag) == 1 {
		panic("loop already finished!, don't reuse it")
	}
	for {
		_ = this.poller.Poll(-1)
		if atomic.LoadInt32(&this.stopFlag) == 1 {
			break
		}
		if ctx == nil {
			continue
		}
		select {
		case <-ctx.Done():
			goto exitLoop
		default:
		}

	}
exitLoop:
	this.processPending()
	close(this.endRunSig)
}
