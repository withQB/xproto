package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"net/http"

	"github.com/gorilla/websocket"
	"github.com/withqb/xproto/connections"
	"github.com/withqb/xproto/multicast"
	"github.com/withqb/xproto/router"
	"github.com/withqb/xproto/util"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	listentcp := flag.String("listen", ":0", "address to listen for TCP connections")
	listenws := flag.String("listenws", ":0", "address to listen for WebSockets connections")
	connect := flag.String("connect", "", "peers to connect to")
	secretkey := flag.String("secretkey", "", "hexadecimal encoded ed25519 key")
	manhole := flag.Bool("manhole", false, "enable the manhole (requires WebSocket listener to be active)")
	flag.Parse()

	var sk ed25519.PrivateKey
	if len(*secretkey) != 0 {
		secretkeyHex, err := hex.DecodeString(*secretkey)
		if err != nil {
			panic(err)
		}

		sk = ed25519.NewKeyFromSeed(secretkeyHex)
	} else {
		var err error
		_, sk, err = ed25519.GenerateKey(nil)
		if err != nil {
			panic(err)
		}
	}

	logger := log.New(os.Stdout, "", 0)
	if hostPort := os.Getenv("PPROFLISTEN"); hostPort != "" {
		logger.Println("Starting pprof on", hostPort)
		go func() {
			_ = http.ListenAndServe(hostPort, nil)
		}()
	}

	listener := net.ListenConfig{}

	xprotoRouter := router.NewRouter(logger, sk, router.RouterOptionBlackhole(true))
	xprotoMulticast := multicast.NewMulticast(logger, xprotoRouter)
	xprotoMulticast.Start()
	xprotoManager := connections.NewConnectionManager(xprotoRouter, nil)

	if connect != nil && *connect != "" {
		for _, uri := range strings.Split(*connect, ",") {
			xprotoManager.AddPeer(strings.TrimSpace(uri))
		}
	}

	if listenws != nil && *listenws != "" {
		go func() {
			var upgrader = websocket.Upgrader{}
			http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					log.Println(err)
					return
				}

				if _, err := xprotoRouter.Connect(
					util.WrapWebSocketConn(conn),
					router.ConnectionURI(conn.RemoteAddr().String()),
					router.ConnectionPeerType(router.PeerTypeRemote),
					router.ConnectionZone("websocket"),
				); err != nil {
					fmt.Println("Inbound WS connection", conn.RemoteAddr(), "error:", err)
					_ = conn.Close()
				} else {
					fmt.Println("Inbound WS connection", conn.RemoteAddr(), "is connected")
				}
			})

			if *manhole {
				fmt.Println("Enabling manhole on HTTP listener")
				http.DefaultServeMux.HandleFunc("/manhole", func(w http.ResponseWriter, r *http.Request) {
					xprotoRouter.ManholeHandler(w, r)
				})
			}

			listener, err := listener.Listen(context.Background(), "tcp", *listenws)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Listening for WebSockets on http://%s\n", listener.Addr())

			if err := http.Serve(listener, http.DefaultServeMux); err != nil {
				panic(err)
			}
		}()
	}

	if listentcp != nil && *listentcp != "" {
		go func() {
			listener, err := listener.Listen(context.Background(), "tcp", *listentcp)
			if err != nil {
				panic(err)
			}

			fmt.Println("Listening on", listener.Addr())

			for {
				conn, err := listener.Accept()
				if err != nil {
					panic(err)
				}

				if _, err := xprotoRouter.Connect(
					conn,
					router.ConnectionURI(conn.RemoteAddr().String()),
					router.ConnectionPeerType(router.PeerTypeRemote),
				); err != nil {
					fmt.Println("Inbound TCP connection", conn.RemoteAddr(), "error:", err)
					_ = conn.Close()
				} else {
					fmt.Println("Inbound TCP connection", conn.RemoteAddr(), "is connected")
				}
			}
		}()
	}

	<-sigs
}
