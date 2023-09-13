package main

import (
	"crypto/ed25519"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"net/http"

	"github.com/withqb/xproto/cmd/xprotoip/tun"
	"github.com/withqb/xproto/connections"
	"github.com/withqb/xproto/multicast"
	"github.com/withqb/xproto/router"
)

func main() {
	_, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	listen := flag.String("listen", "", "address to listen on")
	connect := flag.String("connect", "", "peer to connect to")
	flag.Parse()

	addr, err := net.ResolveTCPAddr("tcp", *listen)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "", 0)
	if hostPort := os.Getenv("PPROFLISTEN"); hostPort != "" {
		go func() {
			listener, err := net.Listen("tcp", hostPort)
			if err != nil {
				panic(err)
			}
			logger.Println("Starting pprof on", listener.Addr())
			if err := http.Serve(listener, nil); err != nil {
				panic(err)
			}
		}()
	}

	xprotoRouter := router.NewRouter(logger, sk)
	xprotoMulticast := multicast.NewMulticast(logger, xprotoRouter)
	xprotoMulticast.Start()
	xprotoManager := connections.NewConnectionManager(xprotoRouter, nil)
	xprotoTUN, err := tun.NewTUN(xprotoRouter)
	if err != nil {
		panic(err)
	}
	_ = xprotoTUN

	if connect != nil && *connect != "" {
		xprotoManager.AddPeer(*connect)
	}

	go func() {
		fmt.Println("Listening on", listener.Addr())

		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				panic(err)
			}

			port, err := xprotoRouter.Connect(
				conn,
				router.ConnectionURI(conn.RemoteAddr().String()),
				router.ConnectionPeerType(router.PeerTypeRemote),
			)
			if err != nil {
				panic(err)
			}

			fmt.Println("Inbound connection", conn.RemoteAddr(), "is connected to port", port)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		switch <-sigs {
		case syscall.SIGUSR1:
			fn := fmt.Sprintf("/tmp/profile.%d", os.Getpid())
			logger.Println("Requested profile:", fn)
			fp, err := os.Create(fn)
			if err != nil {
				logger.Println("failed to create profile:", err)
				return
			}
			defer fp.Close()
			if err := pprof.StartCPUProfile(fp); err != nil {
				logger.Println("failed to start profiling:", err)
				return
			}
			time.AfterFunc(time.Second*10, func() {
				pprof.StopCPUProfile()
				logger.Println("Profile written:", fn)
			})
		}
	}
}
