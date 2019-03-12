package socketpair

import (
	"net"
	"os"
	"syscall"
	"time"
)

// PacketSocketPair returns two bidirectionally connected PacketConns.
func PacketSocketPair() (net.PacketConn, net.PacketConn, error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return nil, nil, err
	}

	if err := syscall.SetNonblock(int(fds[0]), true); err != nil {
		return nil, nil, err
	}
	if err := syscall.SetNonblock(int(fds[1]), true); err != nil {
		return nil, nil, err
	}

	f1 := os.NewFile(uintptr(fds[0]), "socket pair end 0")
	sc1, err := f1.SyscallConn()
	if err != nil {
		return nil, nil, err
	}

	f2 := os.NewFile(uintptr(fds[1]), "socket pair end 1")
	sc2, err := f2.SyscallConn()
	if err != nil {
		return nil, nil, err
	}

	n1 := &socketPair{
		f:  f1,
		rc: sc1,
	}
	n2 := &socketPair{
		f:  f2,
		rc: sc2,
	}
	return n1, n2, nil
}

type socketPair struct {
	f  *os.File
	rc syscall.RawConn
}

func (s *socketPair) LocalAddr() net.Addr {
	return nil
}

func (s *socketPair) SetDeadline(t time.Time) error {
	return s.f.SetDeadline(t)
}

func (s *socketPair) SetReadDeadline(t time.Time) error {
	return s.f.SetReadDeadline(t)
}

func (s *socketPair) SetWriteDeadline(t time.Time) error {
	return s.f.SetWriteDeadline(t)
}

func (s *socketPair) Close() (err error) {
	return s.f.Close()
}

func (s *socketPair) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	cerr := s.rc.Read(func(fd uintptr) bool {
		n, err = syscall.Read(int(fd), p)
		return err != syscall.EAGAIN
	})
	if err != nil {
		return n, nil, err
	}
	return n, nil, cerr
}

func (s *socketPair) WriteTo(p []byte, _ net.Addr) (n int, err error) {
	cerr := s.rc.Write(func(fd uintptr) bool {
		n, err = syscall.Write(int(fd), p)
		return err != syscall.EAGAIN
	})
	if err != nil {
		return n, err
	}
	return 0, cerr
}
