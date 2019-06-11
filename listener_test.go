package liblpc

import (
	"fmt"
	"liblpc/backend"
	"net"
	"os"
	"testing"
	"time"
)

func onStream(sw StreamWriter, data []byte, len int, err error) {
	if err != nil {
		backend.CloseIgnoreErr(sw)
		fmt.Println("onread got error ", err)
		sw.(*FdStream).Loop().Break()
		return
	}
	fmt.Println("onread", string(data))
}

func onAccept(ln *FdListener, newFd int) {
	fmt.Println("on accept , newfd = ", newFd)
	NewFdStream(ln.Loop().(*IOEvtLoop), newFd, onStream)
}

func localConTester(addr *net.UnixAddr) {
	conn, e := net.DialUnix("unix", nil, addr)
	backend.PanicIfError(e)
	defer backend.CloseIgnoreErr(conn)
	tm := time.After(time.Second * 10)
	for {
		select {
		case <-tm:
			return
		default:
		}
		time.Sleep(time.Millisecond * 500)
		conn.Write([]byte(fmt.Sprintf("hello , time -> %s", time.Now().String())))
	}
}

func TestListener(t *testing.T) {
	_ = os.Remove("test_xyz.ipc")
	defer os.Remove("test_xyz.ipc")
	loop, err := NewIOEvtLoop(1024 * 4)
	backend.PanicIfError(err)
	defer backend.CloseIgnoreErr(loop)
	addr, err := net.ResolveUnixAddr("unix", "test_xyz.ipc")
	backend.PanicIfError(err)
	listener, err := net.ListenUnix("unix", addr)
	backend.PanicIfError(err)
	f, err := listener.File()
	backend.PanicIfError(err)
	fdl := NewFdListener(loop, int(f.Fd()), onAccept)
	defer backend.CloseIgnoreErr(fdl)
	go localConTester(addr)
	loop.Run()
}
