package backend

import "container/list"

type EventLoop interface {
	RunInLoop(cb func())
	Notify()
	Run()
	Stop()
	Poller() Poller
}

type evtLoop struct {
	poller Poller
	notify *NotifyWatcher
	cbQ    *list.List
	lock   *SpinLock
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
	return l, nil
}

func (this *evtLoop) RunInLoop(cb func()) {
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
	for ; ls.Len() != 0; {
		front := ls.Front()
		val := front.Value.(func())
		ls.Remove(front)
		val()
	}
}

func (this *evtLoop) Stop() {

}

func (this *evtLoop) Poller() Poller {
	return this.poller
}

func (this *evtLoop) Run() {
	for {
		_ = this.poller.Poll(-1)
	}
	this.processPending()
}
