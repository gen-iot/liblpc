// +build linux

package liblpc

import "C"
import (
	"gitee.com/Puietel/std"
	"net"
	"syscall"
)

type SockFd int
type Fd int

//create new socket , cloexec by default
func NewTcpSocketFd(version int, nonblock bool, cloexec bool) (SockFd, error) {
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
	tp := syscall.SOCK_STREAM
	if nonblock {
		tp |= syscall.SOCK_NONBLOCK
	}
	if cloexec {
		tp |= syscall.SOCK_CLOEXEC
	}
	fd, err := syscall.Socket(domainType, tp, syscall.IPPROTO_TCP)

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

// best way to set cloexec
func (this Fd) Cloexec(enable bool) error {
	syscall.ForkLock.Lock()
	defer syscall.ForkLock.Unlock()
	flags, err := this.FcntlGetFlag()
	if err != nil {
		return err
	}
	return this.FcntlSetFlag(flags | syscall.FD_CLOEXEC)
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

// fd with nonblock, cloexec default
func NewListenerFd(version int, sockAddr syscall.Sockaddr, backLog int, reuseAddr, reusePort bool) (SockFd, error) {
	fd, err := NewTcpSocketFd(version, true, true)
	if err != nil {
		return -1, err
	}
	if err = fd.ReuseAddr(reuseAddr); err != nil {
		return -1, err
	}
	if err = fd.ReusePort(reusePort); err != nil {
		return -1, err
	}
	if err = fd.Bind(sockAddr); err != nil {
		return -1, err
	}
	if err = fd.Listen(backLog); err != nil {
		return -1, err
	}
	return SockFd(fd), nil
}

func NewConnFd(version int, sockAddr syscall.Sockaddr) (SockFd, error) {
	fd, err := NewTcpSocketFd(version, true, true)
	if err != nil {
		return -1, err
	}
	err = fd.Connect(sockAddr)
	if err == nil {
		return fd, nil
	}
	errno, ok := err.(syscall.Errno)
	std.Assert(ok, "unknown err type")
	if errno == syscall.EINPROGRESS || WOULDBLOCK(errno) {
		return fd, nil
	}
	return -1, err
}

func NewConnFdSimple(addrS string) (SockFd, error) {
	sockAddr, err := ResolveTcpAddrSimple(addrS)
	if err != nil {
		return -1, err
	}
	return NewConnFd(sockAddr.Version, sockAddr.Sockaddr)
}

type UnknownAFError string

func (e UnknownAFError) Error() string { return "unknown Addr Type " + string(e) }

type SyscallSockAddr struct {
	syscall.Sockaddr
	Version int
}

func ResolveTcpAddrSimple(addrS string) (*SyscallSockAddr, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addrS)
	if err != nil {
		return nil, err
	}
	if ip4 := tcpAddr.IP.To4(); ip4 != nil {
		return &SyscallSockAddr{
			Sockaddr: &syscall.SockaddrInet4{
				Port: tcpAddr.Port,
				Addr: [4]byte{
					ip4[0],
					ip4[1],
					ip4[2],
					ip4[3]},
			},
			Version: 4,
		}, nil
	}
	if ip6 := tcpAddr.IP.To16(); ip6 != nil {
		sockAddr := &syscall.SockaddrInet6{
			Port: tcpAddr.Port,
			Addr: [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
		for i, b := range ip6 {
			sockAddr.Addr[i] = b
		}
		nic, err := net.InterfaceByName(tcpAddr.Zone)
		if err == nil {
			sockAddr.ZoneId = uint32(nic.Index)
		}
		addr := &SyscallSockAddr{
			Sockaddr: sockAddr,
			Version:  6,
		}
		return addr, nil
	}
	return nil, UnknownAFError(addrS)
}

func ResolveTcpAddr(addrS string) (syscall.Sockaddr, error) {
	sockAddr, err := ResolveTcpAddrSimple(addrS)
	if err != nil {
		return nil, err
	}
	return sockAddr.Sockaddr, err
}
