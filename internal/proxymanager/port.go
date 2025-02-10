// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package proxymanager

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/almeidapaulopt/tsdproxy/internal/core"
	"github.com/almeidapaulopt/tsdproxy/internal/model"

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

func ProviderUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO:
		next.ServeHTTP(w, r)
	})
}

func newPortProxy(ctx context.Context, pconfig model.PortConfig, log zerolog.Logger, accessLog bool) *port {
	log = log.With().Str("port", pconfig.String()).Logger()

	ctxPort, cancel := context.WithCancel(ctx)

	// Create the reverse proxy
	//
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !pconfig.TLSValidate}, //nolint
	}
	reverseProxy := &httputil.ReverseProxy{
		Transport: tr,
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(pconfig.GetFirstTarget())
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

	handler = ProviderUserMiddleware(handler)

	// main http Server
	httpServer := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: core.ReadHeaderTimeout,
		BaseContext:       func(net.Listener) context.Context { return ctxPort },
	}

	return &port{
		log:        log,
		ctx:        ctxPort,
		cancel:     cancel,
		httpServer: httpServer,
	}
}

func newPortRedirect(ctx context.Context, pconfig model.PortConfig, log zerolog.Logger) *port {
	log = log.With().Str("port", pconfig.String()).Logger()

	ctxPort, cancel := context.WithCancel(ctx)

	redirectHTTPServer := &http.Server{
		ReadHeaderTimeout: core.ReadHeaderTimeout,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, pconfig.GetFirstTarget().String(), http.StatusMovedPermanently)
		}),
	}

	return &port{
		log:        log,
		ctx:        ctxPort,
		cancel:     cancel,
		httpServer: redirectHTTPServer,
	}
}

// reverseProxyFunc func is a method that returns a reverse proxy handler.
func reverseProxyFunc(p *httputil.ReverseProxy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.ServeHTTP(w, r)
	})
}

func (p *port) startWithListener(l net.Listener) error {
	p.mtx.Lock()
	p.listener = l
	p.mtx.Unlock()

	err := p.httpServer.Serve(l)
	defer p.log.Info().Msg("Terminating server")

	if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error starting port %w", err)
	}
	return nil
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
