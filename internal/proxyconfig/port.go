// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package proxyconfig

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type (
	PortConfig struct {
		ProxyProtocol  string        `validate:"string" yaml:"proxyProtocol"`
		RedirectURL    *url.URL      `yaml:"redirectUrl"`
		TargetProtocol string        `validate:"string" yaml:"targetProtocol"`
		ProxyPort      int           `validate:"hostname_port" yaml:"proxyPort"`
		TargetPort     int           `validate:"hostname_port" yaml:"targetPort"`
		IsRedirect     bool          `validate:"boolean" yaml:"isRedirect"`
		TLSValidate    bool          `validate:"boolean" yaml:"tlsValidate"`
		Tailscale      TailscalePort `validate:"dive" yaml:"tailscale"`
	}

	TailscalePort struct {
		Funnel bool `validate:"boolean" yaml:"funnel"`
	}
)

const (
	redirectSeparator = "->"
	proxySeparator    = ":"
	protocolSeparator = "/"
)

var (
	ErrInvalidPortFormat   = errors.New("invalid format, missing '" + protocolSeparator + "' or '" + redirectSeparator + "'")
	ErrInvalidProxyConfig  = errors.New("invalid proxy configuration")
	ErrInvalidTargetConfig = errors.New("invalid target configuration")
)

// NewPort parses a port configuration string and returns a PortConfig struct.
//
// The input string `s` must follow one of these formats:
// 1. "<proxy port>/<proxy protocol>:<target port>/<target protocol>"
//   - Example: "443/https:80/http"
//
// 2. "<proxy port>:<target port>"
//   - Example: "443:80"
//   - Defaults: "https" for `proxy protocol` and "http" for `target protocol`.
//
// 3. "<proxy port>/<proxy protocol>-><target URL>"
//   - Example: "443/https->https://example.com"
//   - This format indicates a redirect, setting `IsRedirect` to true and TargetURL.
//
// Returns:
// - PortConfig: A struct containing parsed proxy and target configurations.
// - error: An error if the input string is invalid.
//
// Examples:
// 1. "443/https:80/http" -> ProxyPort=443, ProxyProtocol="https", TargetPort=80, TargetProtocol="http"
// 2. "443:80" -> ProxyPort=443, ProxyProtocol="https", TargetPort=80, TargetProtocol="http"
// 3. "443/https->https://example.com" -> ProxyPort=443, ProxyProtocol="https", IsRedirect=true, TargetURL=https://example.com
func NewPort(s string) (PortConfig, error) {
	config := defaultPortConfig()

	separator, isRedirect := detectSeparator(s)
	config.IsRedirect = isRedirect

	parts := strings.Split(s, separator)
	if len(parts) != 2 { //nolint:mnd
		return config, ErrInvalidProxyConfig
	}

	err := parseProxySegment(parts[0], &config)
	if err != nil {
		return config, err
	}

	if config.IsRedirect {
		err = parseRedirectTarget(parts[1], &config)
	} else {
		err = parseTargetSegment(parts[1], &config)
	}

	return config, err
}

// defaultPortConfig initializes a PortConfig with default values.
func defaultPortConfig() PortConfig {
	return PortConfig{
		ProxyProtocol:  "https",
		TargetProtocol: "http",
		ProxyPort:      443, //nolint:mnd
		TargetPort:     80,  //nolint:mnd
		IsRedirect:     false,
	}
}

// detectSeparator determines the separator used in the configuration string and whether it's a redirect.
func detectSeparator(s string) (string, bool) {
	if strings.Contains(s, redirectSeparator) {
		return redirectSeparator, true
	}
	return proxySeparator, false
}

// parseProxySegment parses the proxy segment of the configuration string.
func parseProxySegment(segment string, config *PortConfig) error {
	proxyParts := strings.Split(segment, protocolSeparator)
	if len(proxyParts) > 2 { //nolint:mnd
		return ErrInvalidProxyConfig
	}

	proxyPort, err := strconv.Atoi(proxyParts[0])
	if err != nil {
		return fmt.Errorf("invalid proxy port: %w", err)
	}
	config.ProxyPort = proxyPort

	if len(proxyParts) == 2 { //nolint:mnd
		config.ProxyProtocol = proxyParts[1]
	}

	return nil
}

func parseTargetSegment(segment string, config *PortConfig) error {
	targetParts := strings.Split(segment, protocolSeparator)
	if len(targetParts) > 2 { //nolint:mnd
		return ErrInvalidTargetConfig
	}

	targetPort, err := strconv.Atoi(targetParts[0])
	if err != nil {
		return fmt.Errorf("invalid target port: %w", err)
	}
	config.TargetPort = targetPort

	if len(targetParts) == 2 { //nolint:mnd
		config.TargetProtocol = targetParts[1]
	}

	return nil
}

func parseRedirectTarget(segment string, config *PortConfig) error {
	targetURL, err := url.Parse(segment)
	if err != nil || targetURL.Scheme == "" || targetURL.Host == "" {
		return fmt.Errorf("invalid target URL: %v", segment)
	}

	config.RedirectURL = targetURL
	config.TargetProtocol = ""
	config.TargetPort = 0
	return nil
}
