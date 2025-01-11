// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/almeidapaulopt/tsdproxy/internal/proxyconfig"
	"github.com/almeidapaulopt/tsdproxy/internal/proxyproviders"

	"github.com/rs/zerolog"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"
	"tailscale.com/tsnet"
)

// Proxy struct implements proxyconfig.Proxy.
type Proxy struct {
	log      zerolog.Logger
	config   *proxyconfig.Config
	tsServer *tsnet.Server
	lc       *tailscale.LocalClient
	ctx      context.Context

	events chan proxyproviders.ProxyEvent

	authURL string
	url     string
	state   proxyconfig.ProxyState

	mtx sync.RWMutex
}

var _ proxyproviders.ProxyInterface = (*Proxy)(nil)

// Start method implements proxyconfig.Proxy Start method.
func (p *Proxy) Start(ctx context.Context) error {
	var (
		err error
		lc  *tailscale.LocalClient
	)

	if err = p.tsServer.Start(); err != nil {
		return err
	}

	if lc, err = p.tsServer.LocalClient(); err != nil {
		return err
	}

	p.mtx.Lock()
	p.ctx = ctx
	p.lc = lc
	p.mtx.Unlock()

	go p.watchStatus()

	return nil
}

func (p *Proxy) GetURL() string {
	// TODO: should be configurable and not force to https
	return "https://" + p.url
}

func (p *Proxy) watchStatus() {
	watcher, err := p.lc.WatchIPNBus(p.ctx, ipn.NotifyInitialState|ipn.NotifyNoPrivateKeys|ipn.NotifyInitialHealthState)
	if err != nil {
		p.log.Error().Err(err).Msg("tailscale.watchStatus")
		return
	}
	defer watcher.Close()

	for {
		n, err := watcher.Next()
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				p.log.Error().Err(err).Msg("tailscale.watchStatus: Next")
			}
			return
		}

		if n.ErrMessage != nil {
			p.log.Error().Str("error", *n.ErrMessage).Msg("tailscale.watchStatus: backend")
			return
		}

		status, err := p.lc.Status(p.ctx)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			p.log.Error().Err(err).Msg("tailscale.watchStatus: status")
			return
		}

		switch status.BackendState {
		case "NeedsLogin":
			p.setState(proxyconfig.ProxyStateAuthenticating, "", status.AuthURL)
		case "Starting":
			p.setState(proxyconfig.ProxyStateStarting, "", "")
		case "Running":
			p.setState(proxyconfig.ProxyStateRunning, strings.TrimRight(status.Self.DNSName, "."), "")
			if p.state != proxyconfig.ProxyStateRunning {
				p.getTLSCertificates()
			}
		}
	}
}

func (p *Proxy) setState(state proxyconfig.ProxyState, url string, authURL string) {
	if p.state == state && p.url == url && p.authURL == authURL {
		return
	}

	p.log.Debug().Str("authURL", url).Str("state", state.String()).Msg("tailscale status")

	p.mtx.Lock()
	p.state = state
	if url != "" {
		p.url = url
	}
	if authURL != "" {
		p.authURL = authURL
	}
	p.mtx.Unlock()

	p.events <- proxyproviders.ProxyEvent{
		State: state,
	}
}

// Close method implements proxyconfig.Proxy Close method.
func (p *Proxy) Close() error {
	close(p.events)
	if p.tsServer != nil {
		return p.tsServer.Close()
	}

	return nil
}

// NewListener method implements proxyconfig.Proxy NewListener method.
func (p *Proxy) NewListener(network, addr string) (net.Listener, error) {
	return p.tsServer.Listen(network, addr)
}

// NewTLSListener method implements proxyconfig.Proxy NewTLSListener method.
func (p *Proxy) NewTLSListener(network, addr string) (net.Listener, error) {
	if p.config.Tailscale.Funnel {
		return p.tsServer.ListenFunnel(network, addr)
	}

	return p.tsServer.ListenTLS(network, addr)
}

func (p *Proxy) WatchEvents() chan proxyproviders.ProxyEvent {
	return p.events
}

func (p *Proxy) GetAuthURL() string {
	return p.authURL
}

func (p *Proxy) getTLSCertificates() {
	p.log.Info().Msg("Generating TLS certificate")
	certDomains := p.tsServer.CertDomains()
	if _, _, err := p.lc.CertPair(p.ctx, certDomains[0]); err != nil {
		p.log.Error().Err(err).Msg("error to get TLS certificates")
		return
	}
	p.log.Info().Msg("TLS certificate generated")
}
