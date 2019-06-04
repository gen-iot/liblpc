package liblpc

import (
	"testing"
	"time"
)

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func testEvtloop(evtLoop *EvtLoop) {
	time.Sleep(time.Second)
	for {
		time.Sleep(time.Second)
		evtLoop.Notify()
	}
}

func TestNotify(t *testing.T) {
	evtLoop, err := NewEvtLoop()
	go testEvtloop(evtLoop)
	panicIfError(err)
	evtLoop.Run()
}
