package simulator

import (
	"fmt"
	"net"
	"time"

	"github.com/withqb/xproto/router"
	"github.com/withqb/xproto/util"
)

func (sim *Simulator) ConnectNodes(a, b string) error {
	if a == b {
		return fmt.Errorf("invalid node pair, a node cannot peer with iself")
	}
	sim.nodesMutex.RLock()
	na := sim.nodes[a]
	nb := sim.nodes[b]
	sim.nodesMutex.RUnlock()
	if na == nil || nb == nil {
		return fmt.Errorf("invalid node pair, one or both of the nodes don't exist")
	}

	sim.wiresMutex.RLock()
	wa := sim.wires[a][b]
	wb := sim.wires[b][a]
	sim.wiresMutex.RUnlock()
	if wa != nil || wb != nil {
		return fmt.Errorf("already connected")
	}

	register := func(conn net.Conn) {
		sim.wiresMutex.Lock()
		defer sim.wiresMutex.Unlock()
		if sim.wires[a] == nil {
			sim.wires[a] = map[string]net.Conn{}
		}
		sim.wires[a][b] = conn
	}

	if sim.sockets {
		c, err := net.DialTCP(na.l.Addr().Network(), nil, na.ListenAddr)
		if err != nil {
			return fmt.Errorf("net.Dial: %w", err)
		}
		if err := c.SetNoDelay(true); err != nil {
			panic(err)
		}
		sc := &util.SlowConn{Conn: c, ReadJitter: 5 * time.Millisecond}
		if _, err := nb.Connect(
			sc,
			router.ConnectionKeepalives(true),
			router.ConnectionPeerType(router.PeerTypeRemote),
		); err != nil {
			return fmt.Errorf("nb.AuthenticatedConnect: %w", err)
		}
		register(sc)
	} else {
		pa, pb := net.Pipe()
		pa = &util.SlowConn{Conn: pa, ReadJitter: 1 * time.Millisecond}
		pb = &util.SlowConn{Conn: pb, ReadJitter: 1 * time.Millisecond}
		go func() {
			if _, err := na.Connect(
				pa,
				router.ConnectionPublicKey(nb.PublicKey()),
				router.ConnectionKeepalives(false),
				router.ConnectionPeerType(router.PeerTypePipe),
			); err != nil {
				return
			}
		}()
		go func() {
			if _, err := nb.Connect(
				pb,
				router.ConnectionPublicKey(na.PublicKey()),
				router.ConnectionKeepalives(false),
				router.ConnectionPeerType(router.PeerTypePipe),
			); err != nil {
				return
			}
		}()
		register(pa)
	}

	sim.log.Printf("Connected node %q to node %q\n", a, b)
	return nil
}

func (sim *Simulator) DisconnectNodes(a, b string) error {
	sim.wiresMutex.RLock()
	wire := sim.wires[a][b]
	firstIndex, secondIndex := a, b
	if wire == nil {
		wire = sim.wires[b][a]
		firstIndex, secondIndex = b, a
	}
	sim.wiresMutex.RUnlock()
	if wire == nil {
		return fmt.Errorf("nodes not connected")
	}

	sim.wiresMutex.Lock()
	sim.wires[firstIndex][secondIndex] = nil
	sim.wiresMutex.Unlock()

	return wire.Close()
}

func (sim *Simulator) DisconnectAllPeers(disconnectNode string) {
	sim.wiresMutex.Lock()
	defer sim.wiresMutex.Unlock()

	nodeWires := sim.wires[disconnectNode]
	for i, conn := range nodeWires {
		if conn != nil {
			_ = conn.Close()
			sim.wires[disconnectNode][i] = nil
		}
	}

	for node, peers := range sim.wires {
		for peer, conn := range peers {
			if peer == disconnectNode {
				if conn != nil {
					_ = conn.Close()
					sim.wires[node][peer] = nil
				}
			}
		}
	}
}
