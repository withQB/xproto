package sessions

import (
	"net/http"
	"time"
)

type HTTP struct {
	httpServer    *http.Server
	httpMux       *http.ServeMux
	httpTransport *http.Transport
	httpClient    *http.Client
}

func (q *SessionProtocol) HTTP() *HTTP {
	t := &http.Transport{
		DisableKeepAlives:   true,
		MaxIdleConnsPerHost: -1,
		Dial:                q.Dial,
		DialTLS:             q.DialTLS,
		DialContext:         q.DialContext,
		DialTLSContext:      q.DialTLSContext,
	}

	h := &HTTP{
		httpServer: &http.Server{
			IdleTimeout:  time.Second * 30,
			ReadTimeout:  time.Second * 10,
			WriteTimeout: time.Second * 10,
		},
		httpMux:       &http.ServeMux{},
		httpTransport: t,
	}

	h.httpServer.Handler = h.httpMux
	h.httpClient = &http.Client{
		Transport: t,
		Timeout:   time.Second * 30,
	}

	go h.httpServer.Serve(q) // nolint:errcheck
	return h
}

func (h *HTTP) Mux() *http.ServeMux {
	return h.httpMux
}

func (h *HTTP) Client() *http.Client {
	return h.httpClient
}
