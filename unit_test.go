package liblpc

import (
	"fmt"
	"os"
	"syscall"
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
		evtLoop.RunInLoop(func() {
			fmt.Println(time.Now().String())
		})
	}
}

func TestNotify(t *testing.T) {
	evtLoop, err := NewEvtLoop()
	go testEvtloop(evtLoop)
	panicIfError(err)
	evtLoop.Run()
}

func TestIOEvtLoop(t *testing.T) {
	fds, e := MakeIpcSockpair(true)
	panicIfError(e)
	loop, e := NewIOEvtLoop(4 * 1024)
	panicIfError(e)
	stream := NewFdStream(loop, int(fds[0]),
		func(sw StreamWriter, data []byte, len int, err error) {
			if err == nil {
				fmt.Println("Server onRead , data is -> ", string(data[:len]))
				sw.Write([]byte(time.Now().String()), true)
			} else {
				fmt.Println("Server onRead error -> ", err, " ,closed!")
				_ = sw.Close()
			}
		})
	stream.Update(true)
	go func() {
		for idx := 0; idx < 10; idx++ {
			time.Sleep(time.Second)
			_, err := syscall.Write(fds[1], []byte(time.Now().String()))
			panicIfError(err)
		}
		err := syscall.Close(fds[1])
		panicIfError(err)
	}()
	loop.Run()
}

func TestSpawnIO(t *testing.T) {
	fmt.Println("current pid = ", os.Getpid())
	fds, err := MakeIpcSockpair(true)
	panicIfError(err)
	loop, err := NewIOEvtLoop(2 * 1024 * 1024)
	panicIfError(err)
	cmd, err := Spawn("bin/child", fds[1])
	panicIfError(err)
	fmt.Println("spawn success pid = ", cmd.Process.Pid)
	stream := NewFdStream(loop, int(fds[0]),
		func(sw StreamWriter, data []byte, len int, err error) {
			if err == nil {
				fmt.Println("Server onRead , data is -> ", string(data[:len]))
				sw.Write([]byte(time.Now().String()), true)
			} else {
				fmt.Println("Server onRead error -> ", err, " ,closed!")
				_ = sw.Close()
			}
		})
	stream.Update(true)
	loop.Run()
}
