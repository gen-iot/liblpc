package liblpc

import (
	"fmt"
	"gitee.com/Puietel/std"
	"net"
	"os"
	"testing"
	"time"
)

func onStream(sw StreamWriter, data []byte, len int, err error) {
	if err != nil {
		sw.(*FdStream).Loop().Break()
		std.CloseIgnoreErr(sw)
		fmt.Println("onread got error ", err)
		return
	}
	fmt.Println("onread", string(data[:len]))
}

func onAccept(ln *FdListener, newFd int) {
	fmt.Println("on accept , newfd = ", newFd)
	stream := NewFdStream(ln.Loop().(*IOEvtLoop), newFd, onStream)
	stream.Start()
}

func localConTester(addr *net.UnixAddr) {
	conn, e := net.DialUnix("unix", nil, addr)
	std.AssertError(e, "dial unix")
	defer std.CloseIgnoreErr(conn)
	tm := time.After(time.Second * 10)
	for {
		select {
		case <-tm:
			return
		default:
		}
		time.Sleep(time.Millisecond * 500)
		_, err := conn.Write([]byte(fmt.Sprintf("hello , time -> %s", time.Now().String())))
		std.AssertError(err, "conn write")
	}
}

func TestListener(t *testing.T) {
	_ = os.Remove("test_xyz.ipc")
	defer func() {
		_ = os.Remove("test_xyz.ipc")
	}()
	loop, err := NewIOEvtLoop(1024 * 4)
	std.AssertError(err, "new io eventloop")
	defer std.CloseIgnoreErr(loop)
	addr, err := net.ResolveUnixAddr("unix", "test_xyz.ipc")
	std.AssertError(err, "resolve unix addr")
	listener, err := net.ListenUnix("unix", addr)
	std.AssertError(err, "listen unix")
	f, err := listener.File()
	std.AssertError(err, "get listener file")
	fdl := NewFdListener(loop, int(f.Fd()), onAccept)
	defer std.CloseIgnoreErr(fdl)
	fdl.Start()
	go localConTester(addr)
	loop.Run()
}
