//go:build linux || netbsd || freebsd || openbsd || dragonflybsd
// +build linux netbsd freebsd openbsd dragonflybsd

package multicast

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/unix"
)

func (m *Multicast) multicastStarted() {
}

func (m *Multicast) udpOptions(network string, address string, c syscall.RawConn) error {
	var reuseport error
	control := c.Control(func(fd uintptr) {
		reuseport = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	})

	switch {
	case reuseport != nil:
		return fmt.Errorf("SO_REUSEPORT: %w", reuseport)
	default:
		return control
	}
}

func (m *Multicast) tcpOptions(network string, address string, c syscall.RawConn) error {
	return nil
}
