//go:build !minimal
// +build !minimal

package router

import (
	"encoding/hex"

	"github.com/Arceliar/phony"
	"github.com/withqb/xproto/router/events"
	"github.com/withqb/xproto/types"
)

type NeighbourInfo struct {
	PublicKey types.PublicKey
}

type PeerInfo struct {
	URI       string
	Port      int
	PublicKey string
	PeerType  int
	Zone      string
}

// Subscribe registers a subscriber to this node's events
func (r *Router) Subscribe(ch chan<- events.Event) {
	phony.Block(r, func() {
		r._subscribers[ch] = &phony.Inbox{}
	})
}

func (r *Router) Coords() types.Coordinates {
	return r.state.coords()
}

func (r *Router) Peers() []PeerInfo {
	var infos []PeerInfo
	phony.Block(r.state, func() {
		for _, p := range r.state._peers {
			if p == nil {
				continue
			}
			infos = append(infos, PeerInfo{
				URI:       string(p.uri),
				Port:      int(p.port),
				PublicKey: hex.EncodeToString(p.public[:]),
				PeerType:  int(p.peertype),
				Zone:      string(p.zone),
			})
		}
	})
	return infos
}

func (r *Router) EnableHopLimiting() {
	r._hopLimiting.Store(true)
}

func (r *Router) DisableHopLimiting() {
	r._hopLimiting.Store(false)
}
