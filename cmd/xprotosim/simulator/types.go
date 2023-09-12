package simulator

import (
	"net"
)

type Node struct {
	SimRouter
	l          *net.TCPListener // nolint:structcheck,unused
	ListenAddr *net.TCPAddr
	Type       APINodeType
}

type Distance struct {
	Real     int64
	Observed int64
}

type RouterConfig struct {
	HopLimiting bool
}
