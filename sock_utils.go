// +build linux

package liblpc

import (
	"github.com/gen-iot/std"
	"golang.org/x/sys/unix"
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
		domainType = unix.AF_INET
	case 6:
		domainType = unix.AF_INET6
	default:
		std.Assert(false, "version must be 4 or 6")
	}
	tp := unix.SOCK_STREAM
	if nonblock {
		tp |= unix.SOCK_NONBLOCK
	}
	if cloexec {
		tp |= unix.SOCK_CLOEXEC
	}
	fd, err := unix.Socket(domainType, tp, unix.IPPROTO_TCP)

	return SockFd(fd), err
}

func (this SockFd) ReuseAddr(enable bool) error {
	opv := 1
	if !enable {
		opv = 0
	}
	return unix.SetsockoptInt(int(this), unix.SOL_SOCKET, unix.SO_REUSEADDR, opv)
}

//noinspection GoSnakeCaseUsage
const SO_REUSEPORT = 0x0F

func (this SockFd) ReusePort(enable bool) error {
	opv := 1
	if !enable {
		opv = 0
	}
	return unix.SetsockoptInt(int(this), unix.SOL_SOCKET, SO_REUSEPORT, opv)
}

func (this SockFd) Bind(sockAddr unix.Sockaddr) error {
	return unix.Bind(int(this), sockAddr)
}

func (this SockFd) Listen(backLog int) error {
	return unix.Listen(int(this), backLog)
}

func (this SockFd) Accept(flags int) (nfd int, sa unix.Sockaddr, err error) {
	return unix.Accept4(int(this), flags)
}

func (this SockFd) Connect(addr unix.Sockaddr) error {
	return unix.Connect(int(this), addr)
}

func (this Fd) NoneBlock(enable bool) error {
	return unix.SetNonblock(int(this), enable)
}

// best way to set cloexec
func (this Fd) Cloexec(enable bool) error {
	syscall.ForkLock.Lock()
	defer syscall.ForkLock.Unlock()
	flags, err := this.FcntlGetFlag()
	if err != nil {
		return err
	}
	return this.FcntlSetFlag(flags | unix.FD_CLOEXEC)
}

func (this Fd) FcntlGetFlag() (flags int, err error) {
	r1, _, eNo := unix.Syscall(unix.SYS_FCNTL, uintptr(this), unix.F_GETFL, 0)
	if eNo != 0 {
		return -1, eNo
	}
	return int(r1), err
}

func (this Fd) FcntlSetFlag(flag int) (err error) {
	_, _, eNo := unix.Syscall(unix.SYS_FCNTL, uintptr(this), unix.F_SETFL, uintptr(flag))
	if eNo != 0 {
		return eNo
	}
	return err
}

func (this Fd) Close() error {
	return unix.Close(int(this))
}

// fd with nonblock, cloexec default
func NewListenerFd2(version int, sockAddr unix.Sockaddr, backLog int, reuseAddr, reusePort bool) (SockFd, error) {
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

// fd with nonblock, cloexec default
func NewListenerFd(addrS string, backLog int, reuseAddr, reusePort bool) (SockFd, error) {
	addr, err := ResolveTcpAddr(addrS)
	if err != nil {
		return -1, err
	}
	return NewListenerFd2(addr.Version, addr, backLog, reuseAddr, reusePort)
}

func NewConnFd2(version int, sockAddr unix.Sockaddr) (SockFd, error) {
	fd, err := NewTcpSocketFd(version, true, true)
	if err != nil {
		return -1, err
	}
	err = fd.Connect(sockAddr)
	if err == nil {
		return fd, nil
	}
	errno, ok := err.(unix.Errno)
	std.Assert(ok, "unknown err type")
	if errno == unix.EINPROGRESS || WOULDBLOCK(errno) {
		return fd, nil
	}
	return -1, err
}

func NewConnFd(addrS string) (SockFd, error) {
	sockAddr, err := ResolveTcpAddr(addrS)
	if err != nil {
		return -1, err
	}
	return NewConnFd2(sockAddr.Version, sockAddr)
}

type UnknownAFError string

func (e UnknownAFError) Error() string { return "unknown Addr Type " + string(e) }

type SyscallSockAddr struct {
	unix.Sockaddr
	Version int
}

func ResolveTcpAddr(addrS string) (*SyscallSockAddr, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addrS)
	if err != nil {
		return nil, err
	}
	if ip4 := tcpAddr.IP.To4(); ip4 != nil {
		return &SyscallSockAddr{
			Sockaddr: &unix.SockaddrInet4{
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
		sockAddr := &unix.SockaddrInet6{
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
