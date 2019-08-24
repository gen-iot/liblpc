package liblpc

import (
	"fmt"
	"github.com/gen-iot/std"
	"golang.org/x/sys/unix"
	"os"
	"testing"
	"time"
)

func testEvtloop(evtLoop EventLoop) {
	time.Sleep(time.Second)
	for {
		time.Sleep(time.Second)
		evtLoop.RunInLoop(func() {
			fmt.Println(time.Now().String())
		})
	}
}

func TestNotify(t *testing.T) {
	evtLoop, err := NewEventLoop()
	go testEvtloop(evtLoop)
	std.AssertError(err, "NewEventLoop")
	evtLoop.Run()
}

func TestIOEvtLoop(t *testing.T) {
	fds, e := MakeIpcSockpair(true)
	std.AssertError(e, "MakeIpcSockpair")
	loop, e := NewIOEvtLoop(4 * 1024)
	std.AssertError(e, "NewIOEvtLoop")
	stream := NewConnStream(loop, int(fds[0]),
		func(sw StreamWriter, data []byte, len int) {
			fmt.Println("Server onRead , data is -> ", string(data[:len]))
			sw.Write([]byte(time.Now().String()), true)
		})
	stream.SetOnClose(func(sw StreamWriter, err error) {
		fmt.Println("Server onRead error -> ", err, " ,closed!")
		std.CloseIgnoreErr(sw)
	})
	defer std.CloseIgnoreErr(stream)
	stream.Start()
	go func() {
		for idx := 0; idx < 10; idx++ {
			time.Sleep(time.Second)
			_, err := unix.Write(fds[1], []byte(time.Now().String()))
			std.AssertError(err, "Write")
		}
		err := unix.Close(fds[1])
		std.AssertError(err, "Close")
	}()
	loop.Run()
}

func TestSpawnIO(t *testing.T) {
	fmt.Println("current pid = ", os.Getpid())
	fds, err := MakeIpcSockpair(true)
	std.AssertError(err, "MakeIpcSockpair")
	loop, err := NewIOEvtLoop(2 * 1024 * 1024)
	std.AssertError(err, "NewIOEvtLoop")
	cmd, err := Spawn("bin/child", fds[1])
	std.AssertError(err, "Spawn")
	fmt.Println("spawn success pid = ", cmd.Process.Pid)
	stream := NewConnStream(loop, int(fds[0]),
		func(sw StreamWriter, data []byte, len int) {
			fmt.Println("Server onRead , data is -> ", string(data[:len]))
			sw.Write([]byte(time.Now().String()), true)
		})
	stream.SetOnClose(func(sw StreamWriter, err error) {
		fmt.Println("Server onRead error -> ", err, " ,closed!")
		_ = sw.Close()
	})
	defer func() {
		_ = stream.Close()
	}()
	stream.Start()
	go func() {
		err := cmd.Wait()
		fmt.Println("child exit error -> ", err)

	}()
	loop.Run()
}
