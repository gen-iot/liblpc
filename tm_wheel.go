package liblpc

import (
	"container/list"
	"context"
	"github.com/gen-iot/std"
	"io"
	"time"
)

type BucketEntry interface {
	io.Closer
	GetRefCounter() *int32
}

type Bucket std.Set // type std.Set<BucketEntry>

type TimeWheel struct {
	buckets       *list.List
	partTimeout   uint16
	partCount     uint16
	cachedEntries chan BucketEntry
}

func NewTimeWheel(partTimeout, partCount uint16) *TimeWheel {
	obj := &TimeWheel{
		buckets:       list.New(),
		partTimeout:   partTimeout,
		partCount:     partCount,
		cachedEntries: make(chan BucketEntry, 32),
	}
	obj.buckets.PushBack(std.NewSet())
	return obj
}

func (this *TimeWheel) Entries() chan<- BucketEntry {
	return this.cachedEntries
}

func (this *TimeWheel) onTick() {
	if this.buckets.Len() < int(this.partCount) {
		this.buckets.PushBack(std.NewSet())
		return
	}
	front := this.buckets.Front()
	this.buckets.Remove(front)
	this.buckets.PushBack(std.NewSet())
	bucketSet := front.Value.(*std.Set)
	bucketSet.Foreach(func(it interface{}) {
		entry := it.(BucketEntry)
		counter := entry.GetRefCounter()
		*counter -= 1
		if *counter == 0 {
			std.CloseIgnoreErr(entry)
		}
	})
}

func (this *TimeWheel) onNewEntry(entry BucketEntry) {
	ptr := entry.GetRefCounter()
	if this.buckets.Back().Value.(*std.Set).Add(entry) {
		*ptr += 1
	}
}

func (this *TimeWheel) Execute(ctx context.Context) {
	partTimeout := time.Duration(this.partTimeout) * time.Second
	timer := time.NewTimer(partTimeout)
	defer timer.Stop()
	defer close(this.cachedEntries)
	for {
		select {
		case <-timer.C:
			this.onTick()
			timer.Reset(partTimeout)
		case entry := <-this.cachedEntries:
			this.onNewEntry(entry)
		case <-ctx.Done():
			stdLog("time wheel exit, due to ctx done")
			return
		}
	}
}
