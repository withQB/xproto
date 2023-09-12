//go:build windows
// +build windows

package multicast

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

func (m *Multicast) multicastStarted() {
}

func (m *Multicast) udpOptions(network string, address string, c syscall.RawConn) error {
	var reuseport error
	control := c.Control(func(fd uintptr) {
		reuseport = windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_REUSEADDR, 1)
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
