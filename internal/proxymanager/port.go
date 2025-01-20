// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package proxymanager

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/almeidapaulopt/tsdproxy/internal/core"
	"github.com/almeidapaulopt/tsdproxy/internal/proxyconfig"

	"github.com/rs/zerolog"
)

type port struct {
	log        zerolog.Logger
	ctx        context.Context
	listener   net.Listener
	cancel     context.CancelFunc
	httpServer *http.Server
	mtx        sync.Mutex
}

func newPortProxy(ctx context.Context, pconfig proxyconfig.PortConfig, log zerolog.Logger, accessLog bool) *port {
	log = log.With().Str("port", pconfig.String()).Logger()

	ctx1, cancel := context.WithCancel(ctx)

	// Create the reverse proxy
	//
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !pconfig.TLSValidate}, //nolint
	}
	reverseProxy := &httputil.ReverseProxy{
		Transport: tr,
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(pconfig.RedirectURL)
			r.Out.Host = r.In.Host
			r.Out.Header["X-Forwarded-For"] = r.In.Header["X-Forwarded-For"]
			r.SetXForwarded()
		},
	}

	handler := reverseProxyFunc(reverseProxy)
	// add logger to proxy
	if accessLog {
		handler = core.LoggerMiddleware(log, handler)
	}

	// main http Server
	httpServer := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: core.ReadHeaderTimeout,
		BaseContext:       func(net.Listener) context.Context { return ctx1 },
	}

	return &port{
		log:        log,
		ctx:        ctx1,
		cancel:     cancel,
		httpServer: httpServer,
	}
}

// reverseProxyFunc func is a method that returns a reverse proxy handler.
func reverseProxyFunc(p *httputil.ReverseProxy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.ServeHTTP(w, r)
	})
}

func newPortRedirect(ctx context.Context, pconfig proxyconfig.PortConfig, log zerolog.Logger) *port {
	log = log.With().Str("port", pconfig.String()).Logger()

	ctx1, cancel := context.WithCancel(ctx)

	redirectHTTPServer := &http.Server{
		ReadHeaderTimeout: core.ReadHeaderTimeout,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}),
	}

	return &port{
		log:        log,
		ctx:        ctx1,
		cancel:     cancel,
		httpServer: redirectHTTPServer,
	}
}

func (p *port) startWithListener(l net.Listener) {
	p.mtx.Lock()
	p.listener = l
	p.mtx.Unlock()

	err := p.httpServer.Serve(l)
	defer func() {
		if r := recover(); r != nil {
			p.log.Error().Err(err).Msg("Panic recovered")
		}
	}()
	defer p.log.Info().Msg("Terminating server")

	if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, http.ErrServerClosed) {
		// TODO: Manager port errors
		// p.state.Store(int32(proxyconfig.ProxyStateError))
		p.log.Error().Err(err).Msg("Error starting proxy server")
	}
}

func (p *port) close() error {
	var errs error

	if p.httpServer != nil {
		errs = errors.Join(errs, p.httpServer.Shutdown(p.ctx))
	}

	if p.listener != nil {
		errs = errors.Join(errs, p.listener.Close())
	}

	p.cancel()

	return errs
}
