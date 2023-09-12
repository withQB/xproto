package router

import (
	"net"
	"time"

	"github.com/Arceliar/phony"
	"github.com/withqb/xproto/types"
	"go.uber.org/atomic"
)

// newLocalPeer returns a new local peer. It should only be called once when
// the router is set up.
func (r *Router) newLocalPeer(blackhole bool) *peer {
	peer := &peer{
		router:   r,
		port:     0,
		context:  r.context,
		cancel:   r.cancel,
		conn:     nil,
		zone:     "local",
		peertype: 0,
		public:   r.public,
		started:  *atomic.NewBool(true),
	}
	if !blackhole {
		peer.traffic = newFairFIFOQueue(trafficBuffer, r.log)
	}
	return peer
}

// ReadFrom reads the next packet that was delivered to this node over the
// Xproto network. Only traffic frames will be returned here (not protocol
// frames). The returned address will either be a `types.PublicKey` (if the
// frame was delivered using SNEK routing) or `types.Coordinates` (if the frame
// was delivered using tree routing).
func (r *Router) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	if r.local.traffic == nil {
		<-r.local.context.Done()
		return 0, nil, nil
	}

	var frame *types.Frame
	readDeadline := r._readDeadline.Load()
	select {
	case <-r.local.context.Done():
		r.local.stop(nil)
		return
	case <-time.After(time.Until(readDeadline)):
		return
	case frame = <-r.local.traffic.pop():
		// A protocol packet is ready to send.
		r.local.traffic.ack()
	}

	addr = frame.SourceKey
	n = len(frame.Payload)
	copy(p, frame.Payload)
	return
}

// WriteTo sends a packet into the Xproto network. The packet will be sent
// as a traffic packet. The supplied net.Addr will dictate the method used to
// route the packet — the address should be a `types.PublicKey` for SNEK routing
// or `types.Coordinates` for tree routing. Supplying an unsupported address type
// will result in a `*net.AddrError` being returned.
func (r *Router) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	timer := time.NewTimer(time.Second * 5)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	switch ga := addr.(type) {
	case types.PublicKey:
		frame := getFrame()
		frame.HopLimit = types.MaxHopLimit
		frame.Type = types.TypeTraffic
		frame.DestinationKey = ga
		phony.Block(r.state, func() {
			if cached, ok := r.state._coordsCache[ga]; ok && time.Since(cached.lastSeen) < coordsCacheLifetime {
				frame.Destination = cached.coordinates
			}
		})
		frame.Source = r.state.coords()
		frame.SourceKey = r.public
		frame.Payload = append(frame.Payload[:0], p...)
		frame.Watermark = types.VirtualSnakeWatermark{
			PublicKey: types.FullMask,
			Sequence:  0,
		}
		phony.Block(r.state, func() {
			_ = r.state._forward(r.local, frame)
		})
		return len(p), nil

	default:
		err = &net.AddrError{
			Err:  "unexpected address type",
			Addr: addr.String(),
		}
		return
	}
}

// LocalAddr returns a net.Addr containing the public key of the node for
// SNEK routing.
func (r *Router) LocalAddr() net.Addr {
	return r.PublicKey()
}

// SetDeadline is not implemented.
func (r *Router) SetDeadline(t time.Time) error {
	return nil
}

func (r *Router) SetReadDeadline(t time.Time) error {
	r._readDeadline.Store(t)
	return nil
}

// SetWriteDeadline is not implemented.
func (r *Router) SetWriteDeadline(t time.Time) error {
	return nil
}
