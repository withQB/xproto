//go:build !linux && !darwin && !netbsd && !freebsd && !openbsd && !dragonflybsd && !windows
// +build !linux,!darwin,!netbsd,!freebsd,!openbsd,!dragonflybsd,!windows

package multicast

import (
	"syscall"
)

func (m *Multicast) multicastStarted() {
}

func (m *Multicast) udpOptions(network string, address string, c syscall.RawConn) error {
	return nil
}

func (m *Multicast) tcpOptions(network string, address string, c syscall.RawConn) error {
	return nil
}
