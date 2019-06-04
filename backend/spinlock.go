package backend

import (
	"runtime"
	"sync/atomic"
)

type spinLock struct {
	flag int32
}

func (this *spinLock) Lock() {
	for ; atomic.CompareAndSwapInt32(&this.flag, 0, 1); {
		runtime.Gosched()
	}
}

func (this *spinLock) Unlock() {
	atomic.StoreInt32(&this.flag, 0)
}
