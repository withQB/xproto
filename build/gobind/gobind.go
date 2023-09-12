package gobind

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"io"
	"log"
	"net"
	"sync"
	"time"

	xprotoConnections "github.com/matrix-org/xproto/connections"
	xprotoMulticast "github.com/matrix-org/xproto/multicast"
	xprotoRouter "github.com/matrix-org/xproto/router"
	"github.com/matrix-org/xproto/types"

	_ "golang.org/x/mobile/bind"
)

const (
	PeerTypeRemote    = xprotoRouter.PeerTypeRemote
	PeerTypeMulticast = xprotoRouter.PeerTypeMulticast
	PeerTypeBluetooth = xprotoRouter.PeerTypeBluetooth
	PeerTypeBonjour   = xprotoRouter.PeerTypeBonjour
)

type Xproto struct {
	ctx             context.Context
	cancel          context.CancelFunc
	logger          *log.Logger
	XprotoRouter    *xprotoRouter.Router
	XprotoMulticast *xprotoMulticast.Multicast
	XprotoManager   *xprotoConnections.ConnectionManager
}

func (m *Xproto) PeerCount(peertype int) int {
	return m.XprotoRouter.PeerCount(peertype)
}

func (m *Xproto) PublicKey() string {
	return m.XprotoRouter.PublicKey().String()
}

func (m *Xproto) SetMulticastEnabled(enabled bool) {
	if enabled {
		m.XprotoMulticast.Start()
	} else {
		m.XprotoMulticast.Stop()
		m.DisconnectType(int(xprotoRouter.PeerTypeMulticast))
	}
}

func (m *Xproto) SetStaticPeer(uri string) {
	m.XprotoManager.RemovePeers()
	if uri != "" {
		m.XprotoManager.AddPeer(uri)
	}
}

func (m *Xproto) DisconnectType(peertype int) {
	for _, p := range m.XprotoRouter.Peers() {
		if int(peertype) == p.PeerType {
			m.XprotoRouter.Disconnect(types.SwitchPortID(p.Port), nil)
		}
	}
}

func (m *Xproto) DisconnectZone(zone string) {
	for _, p := range m.XprotoRouter.Peers() {
		if zone == p.Zone {
			m.XprotoRouter.Disconnect(types.SwitchPortID(p.Port), nil)
		}
	}
}

func (m *Xproto) DisconnectPort(port int) {
	m.XprotoRouter.Disconnect(types.SwitchPortID(port), nil)
}

func (m *Xproto) Conduit(zone string, peertype int) (*Conduit, error) {
	l, r := net.Pipe()
	conduit := &Conduit{conn: r, port: 0}
	go func() {
		conduit.portMutex.Lock()
		defer conduit.portMutex.Unlock()
	loop:
		for i := 1; i <= 10; i++ {
			var err error
			conduit.port, err = m.XprotoRouter.Connect(
				l,
				xprotoRouter.ConnectionZone(zone),
				xprotoRouter.ConnectionPeerType(peertype),
			)
			switch err {
			case io.ErrClosedPipe:
				return
			case io.EOF:
				break loop
			case nil:
				return
			default:
				time.Sleep(time.Second)
			}
		}
		_ = l.Close()
		_ = r.Close()
	}()
	return conduit, nil
}

// nolint:gocyclo
func (m *Xproto) Start() {
	pk, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	m.ctx, m.cancel = context.WithCancel(context.Background())

	m.logger = log.New(BindLogger{}, "Xproto: ", 0)
	m.logger.Println("Public key:", hex.EncodeToString(pk))

	m.XprotoRouter = xprotoRouter.NewRouter(m.logger, sk)
	m.XprotoMulticast = xprotoMulticast.NewMulticast(m.logger, m.XprotoRouter)
	m.XprotoManager = xprotoConnections.NewConnectionManager(m.XprotoRouter, nil)
}

func (m *Xproto) Stop() {
	m.XprotoMulticast.Stop()
	_ = m.XprotoRouter.Close()
	m.cancel()
}

const MaxFrameSize = types.MaxFrameSize

type Conduit struct {
	conn      net.Conn
	port      types.SwitchPortID
	portMutex sync.Mutex
}

func (c *Conduit) Port() int {
	c.portMutex.Lock()
	defer c.portMutex.Unlock()
	return int(c.port)
}

func (c *Conduit) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *Conduit) ReadCopy() ([]byte, error) {
	var buf [65535 * 2]byte
	n, err := c.conn.Read(buf[:])
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (c *Conduit) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *Conduit) Close() error {
	return c.conn.Close()
}
