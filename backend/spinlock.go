package backend

import (
	"runtime"
	"sync/atomic"
)

type SpinLock struct {
	flag int32
}

func (this *SpinLock) Lock() {
	for ; !atomic.CompareAndSwapInt32(&this.flag, 0, 1); {
		runtime.Gosched()
	}
}

func (this *SpinLock) Unlock() {
	atomic.StoreInt32(&this.flag, 0)
}
