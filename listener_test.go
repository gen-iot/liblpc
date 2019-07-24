package liblpc

import (
	"fmt"
	"github.com/gen-iot/std"
	"net"
	"testing"
	"time"
)

func onAccept(ln *Listener, newFd int, err error) {
	if err != nil {
		fmt.Println("listener error:", err)
		std.CloseIgnoreErr(ln)
		return
	}
	fmt.Println("on accept , newfd = ", newFd)
	stream := NewConnStream(ln.Loop().(*IOEvtLoop), newFd, func(sw StreamWriter, data []byte, len int) {
		fmt.Println("onread", string(data[:len]))
	})
	stream.SetOnClose(func(sw StreamWriter, err error) {
		fmt.Println("closed! err:", err)
		sw.(*Stream).Loop().Break()
		std.CloseIgnoreErr(sw)
	})
	stream.Start()
}

func localConTester() {
	conn, err := net.Dial("tcp", "127.0.0.1:8001")
	std.AssertError(err, "dial tcp 127.0.0.1:8001 failed")
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
	sockaddr, err := ResolveTcpAddr("127.0.0.1:8001")
	std.AssertError(err, "ResolveTcpAddr [127.0.0.1:8001]")
	sockFd, err := NewListenerFd2(4, sockaddr, 128, true, true)
	ioEvtLoop, err := NewIOEvtLoop(DefaultIOEvtLoopBufferSize)
	std.AssertError(err, "new ioloop")
	defer std.CloseIgnoreErr(ioEvtLoop)
	//
	fdListener := NewListener(ioEvtLoop, int(sockFd), onAccept)
	fdListener.Start()
	//
	cliFd, err := NewConnFd2(4, sockaddr)
	std.AssertError(err, "NewConnFd2")

	fds := NewClientStream(ioEvtLoop, int(cliFd), func(sw StreamWriter, data []byte, len int) {
		fmt.Println("onread", string(data[:len]))
	})
	fds.SetOnConnect(func(sw StreamWriter, err error) {
		if err != nil {
			fmt.Println("connected! err:", err)
			return
		}
		time.AfterFunc(time.Second*3, func() {
			std.CloseIgnoreErr(sw)
		})
		fmt.Println("connect success, will close after 3s")

	})
	fds.SetOnClose(func(sw StreamWriter, err error) {
		fmt.Println("closed! err:", err)
		sw.(*Stream).Loop().Break()
		std.CloseIgnoreErr(sw)
	})
	fds.Start()
	//go localConTester()
	ioEvtLoop.Run()
}
