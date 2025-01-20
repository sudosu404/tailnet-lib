// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package proxyproviders

import (
	"context"
	"net"

	"github.com/almeidapaulopt/tsdproxy/internal/proxyconfig"
)

type (
	// Proxy interface for each proxy provider
	Provider interface {
		NewProxy(cfg *proxyconfig.Config) (ProxyInterface, error)
	}

	// ProxyInterface interface for each proxy
	ProxyInterface interface {
		Start(context.Context) error
		Close() error
		NewListener(network, addr string) (net.Listener, error)
		NewTLSListener(network, addr string) (net.Listener, error)
		GetURL() string
		GetAuthURL() string
		WatchEvents() chan ProxyEvent
	}

	ProxyEvent struct {
		AuthURL string
		Status  proxyconfig.ProxyStatus
	}
)
