// +build linux

package liblpc

import (
	"fmt"
	"github.com/gen-iot/std"
	"net"
	"syscall"
	"testing"
)

func TestResolveTcpAddr(t *testing.T) {
	addr, err := ResolveTcpAddr("192.168.50.1:4040")
	std.AssertError(err, "192.168.50.1:4040")
	fmt.Println(addr)

	_, err = ResolveTcpAddr("192.168.50.100")
	std.Assert(err != nil, "should failed")
}
func TestResolveTcp6Addr(t *testing.T) {
	addr, err := ResolveTcpAddr("[::1]:4040")
	std.AssertError(err, "[::1]:4040")
	fmt.Println(addr)

	addr, err = ResolveTcpAddr("[::1%4]:4040")
	std.AssertError(err, "[::1]:4040")
	fmt.Println(addr)

	addr, err = ResolveTcpAddr("[::1%eth0]:4040")
	std.AssertError(err, "[::1]:4040")
	fmt.Println(addr)

	_, err = ResolveTcpAddr("[::1]")
	std.Assert(err != nil, "should failed")
}

func TestNewTcpSocketFd(t *testing.T) {
	fds := make([]SockFd, 10)
	for idx := 0; idx < 10; idx++ {
		sockFd, err := NewTcpSocketFd(4, false, true)
		std.AssertError(err, "NewTcpSocketFd 4")
		fds[idx] = sockFd
		t.Log("new sock fd -> ", sockFd)
	}
	defer func() {
		for idx := range fds {
			_ = Fd(fds[idx]).Close()
			t.Log(" sock fd  closed -> ", fds[idx])
		}
	}()
}

func TestSockFd_ReuseAddr(t *testing.T) {
	sockFd, err := NewTcpSocketFd(4, false, true)
	std.AssertError(err, "NewTcpSocketFd 4")
	defer std.CloseIgnoreErr(Fd(sockFd))
	std.AssertError(sockFd.ReuseAddr(true), "ReuseAddr true")
	std.AssertError(sockFd.ReuseAddr(false), "ReuseAddr false")
}

func TestSockFd_ReusePort(t *testing.T) {
	sockFd, err := NewTcpSocketFd(4, false, true)
	std.AssertError(err, "NewTcpSocketFd 4")
	defer std.CloseIgnoreErr(Fd(sockFd))
	std.AssertError(sockFd.ReusePort(true), "ReusePort true")
	std.AssertError(sockFd.ReusePort(false), "ReusePort false")
}

func TestSockFd_Bind(t *testing.T) {
	sockFd, err := NewTcpSocketFd(4, false, true)
	std.AssertError(err, "NewTcpSocketFd 4")
	defer std.CloseIgnoreErr(Fd(sockFd))
	addr, err := ResolveTcpAddr("0.0.0.0:8080")
	std.AssertError(err, "ResolveTcpAddr 0.0.0.0:8080")
	std.AssertError(sockFd.Bind(addr), "Bind")
}

func TestSockFd_Listen(t *testing.T) {
	sockFd, err := NewTcpSocketFd(4, false, true)
	std.AssertError(err, "NewTcpSocketFd 4")
	defer std.CloseIgnoreErr(Fd(sockFd))
	addr, err := ResolveTcpAddr("0.0.0.0:8080")
	std.AssertError(err, "ResolveTcpAddr 0.0.0.0:8080")
	std.AssertError(sockFd.Bind(addr), "Bind")
	std.AssertError(sockFd.Listen(128), "Listen")
}

func TestSockFd_Accept(t *testing.T) {
	sockFd, err := NewTcpSocketFd(4, false, true)
	std.AssertError(err, "NewTcpSocketFd 4")
	defer std.CloseIgnoreErr(Fd(sockFd))
	addr, err := ResolveTcpAddr("0.0.0.0:8080")
	std.AssertError(err, "ResolveTcpAddr 0.0.0.0:8080")
	std.AssertError(sockFd.Bind(addr), "Bind")
	std.AssertError(sockFd.Listen(128), "Listen")
	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:8080")
		std.AssertError(err, "dial tcp 127.0.0.1:8080 failed")
		defer std.CloseIgnoreErr(conn)
	}()
	nfd, sa, err := sockFd.Accept(syscall.O_CLOEXEC | syscall.O_NONBLOCK)
	std.AssertError(err, "Accept")
	t.Log("accept success remote addr ->", sa)
	defer std.CloseIgnoreErr(Fd(nfd))
}

func TestSockFd_Connect(t *testing.T) {
	sockFd, err := NewTcpSocketFd(4, false, true)
	std.AssertError(err, "NewTcpSocketFd 4")
	defer std.CloseIgnoreErr(Fd(sockFd))
	addr, err := ResolveTcpAddr("www.baidu.com:80")
	std.AssertError(sockFd.Connect(addr), "Connect")
	t.Log("connect success")
}

func TestFd_NoneBlock(t *testing.T) {
	sockFd, err := NewTcpSocketFd(4, false, true)
	std.AssertError(err, "NewTcpSocketFd 4")
	defer std.CloseIgnoreErr(Fd(sockFd))
	std.AssertError(Fd(sockFd).NoneBlock(true), "NoneBlock true")
	std.AssertError(Fd(sockFd).NoneBlock(false), "NoneBlock false")
}

func TestFd_FcntlGetFlag(t *testing.T) {
	sockFd, err := NewTcpSocketFd(4, false, true)
	std.AssertError(err, "NewTcpSocketFd 4")
	defer std.CloseIgnoreErr(Fd(sockFd))
	std.AssertError(Fd(sockFd).NoneBlock(true), "NoneBlock true")
	flags, err := Fd(sockFd).FcntlGetFlag()
	std.AssertError(err, "FcntlGetFlag 1")
	std.Assert(flags&syscall.O_NONBLOCK != 0, "flag not match")
	std.AssertError(Fd(sockFd).NoneBlock(false), "NoneBlock true")
	flags, err = Fd(sockFd).FcntlGetFlag()
	std.AssertError(err, "FcntlGetFlag 2")
	std.Assert(flags&syscall.O_NONBLOCK == 0, "flag not match")
}

func TestFd_FcntlSetFlag(t *testing.T) {
	sockFd, err := NewTcpSocketFd(4, false, true)
	std.AssertError(err, "NewTcpSocketFd 4")
	defer std.CloseIgnoreErr(Fd(sockFd))

	flags, err := Fd(sockFd).FcntlGetFlag()
	std.AssertError(err, "FcntlGetFlag 1")

	err = Fd(sockFd).FcntlSetFlag(flags | syscall.O_NONBLOCK)
	std.AssertError(err, "FcntlSetFlag O_NONBLOCK")

	flags, err = Fd(sockFd).FcntlGetFlag()
	std.AssertError(err, "FcntlGetFlag 2")

	std.Assert(flags&syscall.O_NONBLOCK != 0, "flag O_NONBLOCK not match")

}
