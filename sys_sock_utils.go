// +build linux

package liblpc

import (
	"gitee.com/Puietel/std"
	"net"
	"syscall"
)

type SockFd int
type Fd int

func NewTcpSocketFd(version int) (SockFd, error) {
	syscall.ForkLock.Lock()
	defer syscall.ForkLock.Unlock()
	domainType := 0
	switch version {
	case 4:
		domainType = syscall.AF_INET
	case 6:
		domainType = syscall.AF_INET6
	default:
		std.Assert(false, "version must be 4 or 6")
	}
	fd, err := syscall.Socket(domainType, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	return SockFd(fd), err
}

func (this SockFd) ReuseAddr(enable bool) error {
	opv := 1
	if !enable {
		opv = 0
	}
	return syscall.SetsockoptInt(int(this), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, opv)
}

//noinspection GoSnakeCaseUsage
const SO_REUSEPORT = 0x0F

func (this SockFd) ReusePort(enable bool) error {
	opv := 1
	if !enable {
		opv = 0
	}
	return syscall.SetsockoptInt(int(this), syscall.SOL_SOCKET, SO_REUSEPORT, opv)
}

func (this SockFd) Bind(sockAddr syscall.Sockaddr) error {
	return syscall.Bind(int(this), sockAddr)
}

func (this SockFd) Listen(backLog int) error {
	return syscall.Listen(int(this), backLog)
}

func (this SockFd) Accept(flags int) (nfd int, sa syscall.Sockaddr, err error) {
	return syscall.Accept4(int(this), flags)
}

func (this SockFd) Connect(addr syscall.Sockaddr) error {
	return syscall.Connect(int(this), addr)
}

func (this Fd) NoneBlock(enable bool) error {
	return syscall.SetNonblock(int(this), enable)
}

func (this Fd) FcntlGetFlag() (flags int, err error) {
	r1, _, eNo := syscall.Syscall(syscall.SYS_FCNTL, uintptr(this), syscall.F_GETFL, 0)
	if eNo != 0 {
		return -1, eNo
	}
	return int(r1), err
}

func (this Fd) FcntlSetFlag(flag int) (err error) {
	_, _, eNo := syscall.Syscall(syscall.SYS_FCNTL, uintptr(this), syscall.F_SETFL, uintptr(flag))
	if eNo != 0 {
		return eNo
	}
	return err
}

func (this Fd) Close() error {
	return syscall.Close(int(this))
}

func ResolveTcpAddr(addrS string) (addr syscall.Sockaddr, err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addrS)
	if err != nil {
		return nil, err
	}
	v4Addr := tcpAddr.IP.To4()
	std.Assert(len(v4Addr) == net.IPv4len, "only support ipv4 addr")
	addr = &syscall.SockaddrInet4{
		Port: tcpAddr.Port,
		Addr: [4]byte{
			v4Addr[0],
			v4Addr[1],
			v4Addr[2],
			v4Addr[3]},
	}
	return
}
