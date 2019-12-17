package liblpc

import (
	"github.com/gen-iot/std"
	"golang.org/x/sys/unix"
	"os"
	"testing"
	"time"
)

func init() {
	Debug = true
}

func testEvtloop(evtLoop EventLoop) {
	time.Sleep(time.Second)
	for {
		time.Sleep(time.Second)
		evtLoop.RunInLoop(func() {
			stdLog(time.Now().String())
		})
	}
}

func TestNotify(t *testing.T) {
	evtLoop, err := NewEventLoop()
	go testEvtloop(evtLoop)
	std.AssertError(err, "NewEventLoop")
	defer std.CloseIgnoreErr(evtLoop)
	evtLoop.Run(nil)
}

func TestEpoll_WatcherCtl_CloseBeforeAdd(t *testing.T) {
	loop, err := NewIOEvtLoop(1024)
	std.AssertError(err, "new io evtloop")
	defer std.CloseIgnoreErr(loop)
	//
	fds, err := MakeIpcSockpair(true)
	std.AssertError(err, "make ipc sockpair .1")
	conn := NewBufferedConnStream(loop, fds[0], func(sw StreamWriter, buf std.ReadableBuffer) {
		stdLog("on read .1")
	})
	conn.Start()
	//
	fds2, err := MakeIpcSockpair(true)
	std.AssertError(err, "make ipc sockpair .2")
	conn2 := NewBufferedConnStream(loop, fds2[0], func(sw StreamWriter, buf std.ReadableBuffer) {
		stdLog("on read .2")
	})
	conn2.Start()

	//note! must close here due to system fd(id) reuse policy
	err = unix.Close(fds[0])
	std.AssertError(err, "close (before add) fds[0] err")

	err = unix.Close(fds[1])
	std.AssertError(err, "close (before add) fds[1] err")

	loop.Run(nil)
}

func TestEpoll_WatcherCtl_CloseAfterAdd(t *testing.T) {
	loop, err := NewIOEvtLoop(1024)
	std.AssertError(err, "new io evtloop")
	defer std.CloseIgnoreErr(loop)
	//
	fds, err := MakeIpcSockpair(true)
	std.AssertError(err, "make ipc sockpair .1")
	conn := NewBufferedConnStream(loop, fds[0], func(sw StreamWriter, buf std.ReadableBuffer) {
		stdLog("on read .1")
	})
	conn.Start()
	//
	fds2, err := MakeIpcSockpair(true)
	std.AssertError(err, "make ipc sockpair .2")
	conn2 := NewBufferedConnStream(loop, fds2[0], func(sw StreamWriter, buf std.ReadableBuffer) {
		stdLog("on read .2")
	})
	conn2.Start()
	go func() {
		time.Sleep(time.Second * 2)
		// note! must close here due to system fd(id) reuse policy
		err = unix.Close(fds[0])
		std.AssertError(err, "close fds[0] err")
		err = conn.Close()
		stdLog("conn close (after add) err:", err)
		err = unix.Close(fds[1])
		std.AssertError(err, "close fds[1] err")
		err = conn2.Close()
		stdLog("conn2 close (after add) err:", err)
	}()
	//
	loop.Run(nil)
}

func TestIOEvtLoop(t *testing.T) {
	fds, e := MakeIpcSockpair(true)
	std.AssertError(e, "MakeIpcSockpair")
	loop, e := NewIOEvtLoop(4 * 1024)
	std.AssertError(e, "NewIOEvtLoop")
	defer std.CloseIgnoreErr(loop)
	stream := NewConnStream(loop, int(fds[0]),
		func(sw StreamWriter, data []byte, len int) {
			stdLog("Server onRead , data is -> ", string(data[:len]))
			sw.Write([]byte(time.Now().String()), true)
		})
	stream.SetOnClose(func(sw StreamWriter, err error) {
		stdLog("Server onRead error -> ", err, " ,closed!")
		std.CloseIgnoreErr(sw)
	})
	defer std.CloseIgnoreErr(stream)
	stream.Start()
	go func() {
		for idx := 0; idx < 100; idx++ {
			time.Sleep(time.Millisecond)
			_, err := unix.Write(fds[1], []byte(time.Now().String()))
			std.AssertError(err, "Write")
		}
		err := unix.Close(fds[1])
		std.AssertError(err, "Close")
		loop.Break()
	}()
	loop.Run(nil)
}

func TestSpawnIO(t *testing.T) {
	stdLog("current pid = ", os.Getpid())
	fds, err := MakeIpcSockpair(true)
	std.AssertError(err, "MakeIpcSockpair")
	loop, err := NewIOEvtLoop(2 * 1024 * 1024)
	std.AssertError(err, "NewIOEvtLoop")
	defer std.CloseIgnoreErr(loop)
	cmd, err := Spawn("bin/child", fds[1])
	std.AssertError(err, "Spawn")
	stdLog("spawn success pid = ", cmd.Process.Pid)
	stream := NewConnStream(loop, int(fds[0]),
		func(sw StreamWriter, data []byte, len int) {
			stdLog("Server onRead , data is -> ", string(data[:len]))
			sw.Write([]byte(time.Now().String()), true)
		})
	stream.SetOnClose(func(sw StreamWriter, err error) {
		stdLog("Server onRead error -> ", err, " ,closed!")
		_ = sw.Close()
	})
	defer func() {
		_ = stream.Close()
	}()
	stream.Start()
	go func() {
		err := cmd.Wait()
		stdLog("child exit error -> ", err)

	}()
	loop.Run(nil)
}
