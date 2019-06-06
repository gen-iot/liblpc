package liblpc

import (
	"container/list"
	"liblpc/backend"
)

type TimeoutRef interface {
	Increase(n int)
}

type Bucket *list.List

type TimeWheel struct {
	bucketQ   *list.List
	bucketNum int
	lock      *backend.SpinLock
}

func NewTimeWheel(bucketNum int) *TimeWheel {
	this := new(TimeWheel)
	this.bucketQ = list.New()
	this.bucketNum = bucketNum
	this.addBucket()
	return this
}

func (this *TimeWheel) addBucket() {
	this.bucketQ.PushBack(list.New())
}

func (this *TimeWheel) popBucket() {
	this.lock.Lock()
	defer this.lock.Unlock()

	front := this.bucketQ.Front()
	ele := front.Value.(*list.List)
	for ; ele.Len() != 0; {
		ef := ele.Front()
		ef.Value.(TimeoutRef).Increase(-1)
		ele.Remove(ef)
	}
	this.bucketQ.Remove(front)
}

func (this *TimeWheel) Add(ref TimeoutRef) {
	this.lock.Lock()
	defer this.lock.Unlock()

	back := this.bucketQ.Back()
	ref.Increase(1)
	back.Value.(*list.List).PushBack(ref)
}

func (this *TimeWheel) Tick() {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.bucketQ.Len() < this.bucketNum {
		this.addBucket()
		return
	}
	this.popBucket()
}
