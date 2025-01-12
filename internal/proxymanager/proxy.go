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
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/almeidapaulopt/tsdproxy/internal/core"
	"github.com/almeidapaulopt/tsdproxy/internal/proxyconfig"
	"github.com/almeidapaulopt/tsdproxy/internal/proxyproviders"

	"github.com/rs/zerolog"
)

type (
	// Proxy struct is a struct that contains all the information needed to run a proxy.
	Proxy struct {
		log                zerolog.Logger
		ctx                context.Context
		providerProxy      proxyproviders.ProxyInterface
		redirectHTTPServer *http.Server
		Config             *proxyconfig.Config
		URL                *url.URL
		reverseProxy       *httputil.ReverseProxy
		httpServer         *http.Server
		cancel             context.CancelFunc
		listeners          []net.Listener
		mtx                sync.RWMutex
		state              atomic.Int32
	}
)

// NewProxy function is a function that creates a new proxy.
func NewProxy(log zerolog.Logger,
	pcfg *proxyconfig.Config,
	proxyProvider proxyproviders.Provider,
) (*Proxy, error) {
	//
	var err error

	log = log.With().Str("proxyname", pcfg.Hostname).Logger()
	log.Info().Str("hostname", pcfg.Hostname).Msg("setting up proxy")

	log.Debug().
		Str("hostname", pcfg.Hostname).
		Str("targetURL", pcfg.TargetURL.String()).
		Str("proxyURL", pcfg.ProxyURL.String()).
		Msg("initializing proxy")

	// Create the reverse proxy
	//
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !pcfg.TLSValidate}, //nolint
	}
	reverseProxy := &httputil.ReverseProxy{
		Transport: tr,
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(pcfg.TargetURL)
			r.Out.Host = r.In.Host
			r.Out.Header["X-Forwarded-For"] = r.In.Header["X-Forwarded-For"]
			r.SetXForwarded()
		},
	}

	// Create the proxyProvider proxy
	//
	pProvider, err := proxyProvider.NewProxy(pcfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing proxy on proxyProvider: %w", err)
	}

	// Create reverse proxy server
	//
	handler := reverseProxyFunc(reverseProxy)

	log.Debug().
		Str("hostname", pcfg.Hostname).
		Msg("Proxy server created successfully")

	// add logger to proxy
	//
	if pcfg.ProxyAccessLog {
		handler = core.LoggerMiddleware(log, handler)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// main http Server
	httpServer := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: core.ReadHeaderTimeout,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	p := &Proxy{
		log: log,
		// targetProvider: targetprovider,
		Config:        pcfg,
		ctx:           ctx,
		cancel:        cancel,
		reverseProxy:  reverseProxy,
		providerProxy: pProvider,
		httpServer:    httpServer,
	}

	return p, nil
}

// Close method is a method that initiate proxy close procedure.
func (proxy *Proxy) Close() {
	proxy.state.Store(int32(proxyconfig.ProxyStateStopping))

	// cancel context
	proxy.cancel()
	// make sure all listeners are closed
	proxy.close()

	proxy.state.Store(int32(proxyconfig.ProxyStateStopped))
}

// close method is a method that closes all listeners ans httpServer.
func (proxy *Proxy) close() {
	var errs error
	proxy.log.Info().Str("name", proxy.Config.Hostname).Msg("stopping proxy")

	if proxy.httpServer != nil {
		if err := proxy.httpServer.Shutdown(proxy.ctx); err != nil {
			proxy.log.Error().Err(err).Msg("Error closing proxy")
		}
	}

	// if has http redirect server
	if proxy.redirectHTTPServer != nil {
		errs = errors.Join(errs, proxy.redirectHTTPServer.Close())
	}
	if proxy.redirectHTTPServer != nil {
		errs = errors.Join(errs, proxy.redirectHTTPServer.Close())
	}
	for _, ls := range proxy.listeners {
		errs = errors.Join(errs, ls.Close())
	}
	if proxy.providerProxy != nil {
		errs = errors.Join(proxy.providerProxy.Close())
	}

	if errs != nil {
		proxy.log.Error().Err(errs).Msg("Error stopping proxy")
	}

	proxy.log.Info().Str("name", proxy.Config.Hostname).Msg("proxy stopped")
}

func (proxy *Proxy) Start() {
	go func() {
		go proxy.start()
		for {
			event := <-proxy.providerProxy.WatchEvents()
			proxy.state.Store(event.State.Int32())
		}
	}()
}

// Start method is a method that starts the proxy.
func (proxy *Proxy) start() {
	proxy.log.Info().Str("name", proxy.Config.Hostname).Msg("starting proxy")

	if err := proxy.providerProxy.Start(proxy.ctx); err != nil {
		proxy.log.Error().Err(err).Msg("Error starting proxy")
		proxy.Close()
		return
	}

	ls, err := proxy.addTLSListener("tcp", ":443")
	if err != nil {
		proxy.log.Error().Err(err).Msg("Error Listening on TLS")
	}

	// Redirect http to https
	err = proxy.startRedirectServer()
	if err != nil {
		proxy.log.Error().Err(err).Msg("Error starting redirect server")
	}

	// start server
	//
	err = proxy.httpServer.Serve(ls)
	defer func() {
		if r := recover(); r != nil {
			proxy.log.Error().Err(err).Msg("Panic recovered")
		}
	}()
	defer proxy.log.Info().Msg("Terminating server")

	if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, http.ErrServerClosed) {
		proxy.state.Store(int32(proxyconfig.ProxyStateError))
		proxy.log.Error().Err(err).Str("hostname", proxy.Config.Hostname).Msg("Error starting proxy server")
		return
	}
}

func (proxy *Proxy) GetState() proxyconfig.ProxyState {
	return proxyconfig.ProxyState(proxy.state.Load())
}

func (proxy *Proxy) addListener(network, addr string) (net.Listener, error) {
	l, err := proxy.providerProxy.NewListener(network, addr)
	if err != nil {
		return nil, err
	}

	proxy.mtx.Lock()
	defer proxy.mtx.Unlock()

	proxy.listeners = append(proxy.listeners, l)

	return l, nil
}

func (proxy *Proxy) addTLSListener(network, addr string) (net.Listener, error) {
	l, err := proxy.providerProxy.NewTLSListener(network, addr)
	if err != nil {
		return nil, err
	}

	proxy.mtx.Lock()
	defer proxy.mtx.Unlock()

	proxy.listeners = append(proxy.listeners, l)

	return l, nil
}

// StartRedirectServer method is a method that starts http rediret server to https.
func (proxy *Proxy) startRedirectServer() error {
	lt, err := proxy.addListener("tcp", ":80")
	if err != nil {
		return fmt.Errorf("error creating redirect HTTP listener: %w", err)
	}

	redirectHTTPServer := &http.Server{
		ReadHeaderTimeout: core.ReadHeaderTimeout,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}),
	}

	proxy.mtx.Lock()
	proxy.redirectHTTPServer = redirectHTTPServer
	proxy.mtx.Unlock()

	go func() {
		err := proxy.redirectHTTPServer.Serve(lt)
		if err != nil && err != http.ErrServerClosed {
			// Log the error, but don't stop the main server
			proxy.log.Error().Err(err).Msg("HTTP redirect server error")
		}
	}()

	return nil
}

func (proxy *Proxy) GetURL() string {
	return proxy.providerProxy.GetURL()
}

func (proxy *Proxy) GetAuthURL() string {
	return proxy.providerProxy.GetAuthURL()
}

// reverseProxyFunc func is a method that returns a reverse proxy handler.
func reverseProxyFunc(p *httputil.ReverseProxy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.ServeHTTP(w, r)
	})
}
