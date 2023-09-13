package sessions

import (
	"net"

	"github.com/quic-go/quic-go"
)

type Stream struct {
	quic.Stream
	connection quic.Connection
}

func (s *Stream) LocalAddr() net.Addr {
	return s.connection.LocalAddr()
}

func (s *Stream) RemoteAddr() net.Addr {
	return s.connection.RemoteAddr()
}
