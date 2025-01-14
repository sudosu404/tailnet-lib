// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package proxyconfig

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
)

type (
	PortConfig struct {
		ProxyProtocol  string        `validate:"string" default:"https" yaml:"proxyProtocol"`
		TargetProtocol string        `validate:"string" default:"http" yaml:"targetProtocol"`
		ProxyPort      int           `validate:"hostname_port" default:"443" yaml:"proxyPort"`
		TargetPort     int           `validate:"hostname_port" default:"80" yaml:"targetPort"`
		IsRedirect     bool          `validate:"boolean" default:"false" yaml:"isRedirect"`
		TLSValidate    bool          `validate:"boolean" default:"true" yaml:"tlsValidate"`
		Tailscale      TailscalePort `validate:"dive" yaml:"tailscale"`
	}

	TailscalePort struct {
		Funnel bool `validate:"boolean" default:"false" yaml:"funnel"`
	}
)

const (
	redirectDelimiter = "->"
	proxyDelimiter    = ":"
	protocolDelimiter = "/"
)

func NewPort(s string) (PortConfig, error) {
	var port PortConfig

	if err := defaults.Set(&port); err != nil {
		return port, fmt.Errorf("error loading defaults: %w", err)
	}

	port.IsRedirect = strings.Contains(s, redirectDelimiter)

	delimiter := proxyDelimiter
	if port.IsRedirect {
		delimiter = redirectDelimiter
	}
	parts := strings.Split(s, delimiter)

	// Parse proxy port and protocol
	proxyParts := strings.Split(parts[0], protocolDelimiter)
	proxyPort, err := strconv.Atoi(proxyParts[0])
	if err != nil {
		return port, fmt.Errorf("invalid proxy port: %w", err)
	}
	port.ProxyPort = proxyPort
	if len(proxyParts) > 1 {
		port.ProxyProtocol = proxyParts[1]
	}

	// Parse target port and protocol
	targetParts := strings.Split(parts[1], protocolDelimiter)
	targetPort, err := strconv.Atoi(targetParts[0])
	if err != nil {
		return port, fmt.Errorf("invalid target port: %w", err)
	}
	port.TargetPort = targetPort
	if len(targetParts) > 1 {
		port.TargetProtocol = targetParts[1]
	}

	return port, nil
}
