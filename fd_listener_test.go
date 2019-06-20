package liblpc

import (
	"fmt"
	"gitee.com/Puietel/std"
	"net"
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
	fdListener := NewFdListener(ioEvtLoop, int(sockFd), onAccept)
	fdListener.Start()
	//
	cliFd, err := NewConnFd2(4, sockaddr)
	std.AssertError(err, "NewConnFd2")

	fds := NewFdStream(ioEvtLoop, int(cliFd), onStream)
	fds.SetOnConnect(func(sw StreamWriter) {
		fmt.Println("connected!")
	})
	fds.Start()
	//go localConTester()
	ioEvtLoop.Run()
}
