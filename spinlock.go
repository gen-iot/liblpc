package liblpc

import (
	"runtime"
	"sync/atomic"
)

const (
	locked   = int32(1)
	unlocked = int32(2)
)

type SpinLock struct {
	flag int32
}

func NewSpinLock() *SpinLock {
	l := new(SpinLock)
	l.flag = unlocked
	return l
}

func NewLockedSpinLock() *SpinLock {
	l := NewSpinLock()
	l.flag = locked
	return l
}

func (this *SpinLock) Lock() {
	for ; !atomic.CompareAndSwapInt32(&this.flag, unlocked, locked); {
		runtime.Gosched()
	}
}

func (this *SpinLock) Unlock() {
	atomic.StoreInt32(&this.flag, unlocked)
}
