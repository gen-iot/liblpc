package liblpc

import (
	"fmt"
	"liblpc/backend"
	"os"
	"syscall"
	"testing"
	"time"
)

func testEvtloop(evtLoop backend.EventLoop) {
	time.Sleep(time.Second)
	for {
		time.Sleep(time.Second)
		evtLoop.RunInLoop(func() {
			fmt.Println(time.Now().String())
		})
	}
}

func TestNotify(t *testing.T) {
	evtLoop, err := backend.NewEventLoop()
	go testEvtloop(evtLoop)
	backend.PanicIfError(err)
	evtLoop.Run()
}

func TestIOEvtLoop(t *testing.T) {
	fds, e := MakeIpcSockpair(true)
	backend.PanicIfError(e)
	loop, e := NewIOEvtLoop(4 * 1024)
	backend.PanicIfError(e)
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
			backend.PanicIfError(err)
		}
		err := syscall.Close(fds[1])
		backend.PanicIfError(err)
	}()
	loop.Run()
}

func TestSpawnIO(t *testing.T) {
	fmt.Println("current pid = ", os.Getpid())
	fds, err := MakeIpcSockpair(true)
	backend.PanicIfError(err)
	loop, err := NewIOEvtLoop(2 * 1024 * 1024)
	backend.PanicIfError(err)
	cmd, err := Spawn("bin/child", fds[1])
	backend.PanicIfError(err)
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
	go func() {
		err := cmd.Wait()
		fmt.Println("child exit error -> ", err)
	}()
	stream.Update(true)
	loop.Run()
}
