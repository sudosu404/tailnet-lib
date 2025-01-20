// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package proxymanager

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"sync"

	"github.com/almeidapaulopt/tsdproxy/internal/proxyconfig"
	"github.com/almeidapaulopt/tsdproxy/internal/proxyproviders"

	"github.com/rs/zerolog"
)

type (
	// Proxy struct is a struct that contains all the information needed to run a proxy.
	Proxy struct {
		log           zerolog.Logger
		ctx           context.Context
		providerProxy proxyproviders.ProxyInterface
		Config        *proxyconfig.Config
		URL           *url.URL
		cancel        context.CancelFunc
		ports         map[string]*port
		mtx           sync.Mutex
		status        proxyconfig.ProxyStatus
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

	log.Debug().Str("hostname", pcfg.Hostname).
		Msg("initializing proxy")

	// Create the proxyProvider proxy
	//
	pProvider, err := proxyProvider.NewProxy(pcfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing proxy on proxyProvider: %w", err)
	}

	log.Debug().
		Str("hostname", pcfg.Hostname).
		Msg("Proxy server created successfully")

	ctx, cancel := context.WithCancel(context.Background())

	p := &Proxy{
		log: log,
		// targetProvider: targetprovider,
		Config:        pcfg,
		ctx:           ctx,
		cancel:        cancel,
		providerProxy: pProvider,
		ports:         make(map[string]*port),
	}

	p.initPorts()

	return p, nil
}

func (proxy *Proxy) Start() {
	go func() {
		proxy.start()
		for {
			event := <-proxy.providerProxy.WatchEvents()
			proxy.setStatus(event.Status)
		}
	}()
}

// Close method is a method that initiate proxy close procedure.
func (proxy *Proxy) Close() {
	proxy.setStatus(proxyconfig.ProxyStatusStopping)

	// cancel context
	proxy.cancel()
	// make sure all listeners are closed
	proxy.close()

	proxy.setStatus(proxyconfig.ProxyStatusStopped)
}

func (proxy *Proxy) setStatus(status proxyconfig.ProxyStatus) {
	proxy.mtx.Lock()
	defer proxy.mtx.Unlock()
	proxy.status = status
}

func (proxy *Proxy) GetStatus() proxyconfig.ProxyStatus {
	proxy.mtx.Lock()
	defer proxy.mtx.Unlock()
	return proxy.status
}

func (proxy *Proxy) GetURL() string {
	return proxy.providerProxy.GetURL()
}

func (proxy *Proxy) GetAuthURL() string {
	return proxy.providerProxy.GetAuthURL()
}

func (proxy *Proxy) initPorts() {
	var newPort *port
	for k, v := range proxy.Config.Ports {
		log := proxy.log.With().Str("port", k).Logger()
		if v.IsRedirect {
			newPort = newPortRedirect(proxy.ctx, v, log)
		} else {
			newPort = newPortProxy(proxy.ctx, v, log, proxy.Config.ProxyAccessLog)
		}

		proxy.log.Debug().Any("port", newPort).Msg("newport")

		proxy.mtx.Lock()
		proxy.ports[k] = newPort
		proxy.mtx.Unlock()
	}
}

func (proxy *Proxy) getListener(protocol string, port int) (net.Listener, error) {
	proto := protocol
	if protocol == "http" || protocol == "https" {
		proto = "tcp"
	}

	addr := ":" + strconv.Itoa(port)

	var l net.Listener
	var err error

	if protocol == "https" {
		l, err = proxy.providerProxy.NewTLSListener(proto, addr)
	} else {
		l, err = proxy.providerProxy.NewListener(proto, addr)
	}
	return l, err
}

// Start method is a method that starts the proxy.
func (proxy *Proxy) start() {
	proxy.log.Info().Str("name", proxy.Config.Hostname).Msg("starting proxy")

	proxy.mtx.Lock()
	portsConfig := proxy.Config.Ports
	portsCount := len(proxy.ports)
	proxy.mtx.Unlock()

	if portsCount == 0 {
		proxy.log.Warn().Msg("No ports configured")
		proxy.setStatus(proxyconfig.ProxyStatusError)

		return
	}

	if err := proxy.providerProxy.Start(proxy.ctx); err != nil {
		proxy.log.Error().Err(err).Msg("Error starting with proxy provider")
		proxy.Close()
		return
	}

	var l net.Listener
	var err error

	for k, v := range portsConfig {
		proxy.log.Debug().Str("port", k).Msg("Starting proxy port")

		l, err = proxy.getListener(v.ProxyProtocol, v.ProxyPort)
		if err != nil {
			proxy.log.Error().Err(err).Str("port", k).Msg("Error adding listener")
		}

		if err := proxy.startPort(k, l); err != nil {
			proxy.log.Error().Err(err).Str("port", k).Msg("Error starting proxy port")
			continue
		}
	}
}

func (proxy *Proxy) startPort(name string, l net.Listener) error {
	proxy.mtx.Lock()
	defer proxy.mtx.Unlock()

	if p, ok := proxy.ports[name]; ok {
		go func() {
			p.setListener(l)
			p.start()
		}()
	}
	return nil
}

// close method is a method that closes all listeners ans httpServer.
func (proxy *Proxy) close() {
	var errs error
	proxy.log.Info().Str("name", proxy.Config.Hostname).Msg("stopping proxy")

	for _, p := range proxy.ports {
		errs = errors.Join(errs, p.close())
	}
	if proxy.providerProxy != nil {
		errs = errors.Join(proxy.providerProxy.Close())
	}

	if errs != nil {
		proxy.log.Error().Err(errs).Msg("Error stopping proxy")
	}

	proxy.log.Info().Str("name", proxy.Config.Hostname).Msg("proxy stopped")
}
