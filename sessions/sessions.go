package sessions

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/withqb/xproto/router"
	"github.com/withqb/xproto/types"
)

type Sessions struct {
	r            *router.Router
	log          types.Logger                // logger
	context      context.Context             // router context
	cancel       context.CancelFunc          // shut down the router
	protocols    map[string]*SessionProtocol // accepted connections by proto
	tlsCert      *tls.Certificate            //
	tlsServerCfg *tls.Config                 //
	quicListener quic.Listener               //
	quicConfig   *quic.Config                //
}

type SessionProtocol struct {
	s         *Sessions
	transport *quic.Transport
	proto     string
	streams   chan net.Conn
	sessions  sync.Map // types.PublicKey -> *activeSession
	closeOnce sync.Once
}

type activeSession struct {
	quic.Connection
	sync.RWMutex
}

func NewSessions(log types.Logger, r *router.Router, protos []string) *Sessions {
	transport := &quic.Transport{Conn: r}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Sessions{
		r:         r,
		log:       log,
		context:   ctx,
		cancel:    cancel,
		protocols: make(map[string]*SessionProtocol, len(protos)),
		quicConfig: &quic.Config{
			MaxIdleTimeout:          time.Second * 15,
			DisablePathMTUDiscovery: true,
		},
	}
	for _, proto := range protos {
		s.protocols[proto] = &SessionProtocol{
			s:         s,
			transport: transport,
			proto:     proto,
			streams:   make(chan net.Conn, 1),
		}
	}

	s.tlsCert = s.generateTLSCertificate()
	s.tlsServerCfg = &tls.Config{
		Certificates: []tls.Certificate{*s.tlsCert},
		ClientAuth:   tls.RequireAnyClientCert,
		NextProtos:   protos,
	}

	listener, err := transport.Listen(s.tlsServerCfg, s.quicConfig)
	if err != nil {
		panic(fmt.Errorf("quic.NewSocketFromPacketConnNoClose: %w", err))
	}
	s.quicListener = *listener

	go s.listener()
	return s
}

func (s *Sessions) Close() error {
	s.cancel()
	return nil
}

func (s *Sessions) Protocol(proto string) *SessionProtocol {
	return s.protocols[proto]
}

func (s *SessionProtocol) Sessions() []ed25519.PublicKey {
	var sessions []ed25519.PublicKey
	s.sessions.Range(func(k, _ interface{}) bool {
		switch pk := k.(type) {
		case types.PublicKey:
			sessions = append(sessions, pk[:])
		default:
		}
		return true
	})
	return sessions
}

func (p *SessionProtocol) getSession(pk types.PublicKey) (*activeSession, bool) {
	v, ok := p.sessions.LoadOrStore(pk, &activeSession{})
	return v.(*activeSession), ok
}

func (s *Sessions) generateTLSCertificate() *tls.Certificate {
	private, public := s.r.PrivateKey(), s.r.PublicKey()
	id := hex.EncodeToString(public[:])

	template := x509.Certificate{
		Subject: pkix.Name{
			CommonName: id,
		},
		SerialNumber: big.NewInt(1),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365),
		DNSNames:     []string{id},
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		ed25519.PublicKey(public[:]),
		ed25519.PrivateKey(private[:]),
	)
	if err != nil {
		panic(fmt.Errorf("x509.CreateCertificate: %w", err))
	}
	privateKey, err := x509.MarshalPKCS8PrivateKey(ed25519.PrivateKey(private[:]))
	if err != nil {
		panic(fmt.Errorf("x509.MarshalPKCS8PrivateKey: %w", err))
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(fmt.Errorf("tls.X509KeyPair: %w", err))
	}

	return &tlsCert
}
