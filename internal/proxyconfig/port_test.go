// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package proxyconfig

import (
	"net/url"
	"testing"
)

func TestNewPortShortLabel(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantConfig PortConfig
		wantErr    bool
	}{
		{
			name:  "Short: valid proxy with protocol",
			input: "443" + protocolSeparator + "https",
			wantConfig: PortConfig{
				ProxyProtocol: "https",
				ProxyPort:     443,
			},
			wantErr: false,
		},
		{
			name:  "Short: valid proxy with without protocol",
			input: "443",
			wantConfig: PortConfig{
				ProxyProtocol: "https",
				ProxyPort:     443,
			},
			wantErr: false,
		},
		{
			name:    "Short: Invalid proxy port",
			input:   "443https",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfig, err := NewPortShortLabel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPortShortLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !compareShortLabelPortConfig(gotConfig, tt.wantConfig) {
				t.Errorf("NewPortShortLabel() = %+v, want %+v", gotConfig, tt.wantConfig)
			}
		})
	}
}

func TestNewPortLongLabel(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantConfig PortConfig
		wantErr    bool
	}{
		{
			name:  "Long: Valid proxy with protocols",
			input: "443" + protocolSeparator + "https" + proxySeparator + "80" + protocolSeparator + "http",
			wantConfig: PortConfig{
				ProxyProtocol:  "https",
				TargetProtocol: "http",
				ProxyPort:      443,
				TargetPort:     80,
				IsRedirect:     false,
			},
			wantErr: false,
		},
		{
			name:  "Long: Valid proxy with only target protocol",
			input: "80" + proxySeparator + "443" + protocolSeparator + "https",
			wantConfig: PortConfig{
				ProxyProtocol:  "https",
				TargetProtocol: "https",
				ProxyPort:      80,
				TargetPort:     443,
				IsRedirect:     false,
			},
			wantErr: false,
		},
		{
			name:  "Long: Valid proxy with only proxy protocol",
			input: "80" + protocolSeparator + "http" + proxySeparator + "8080",
			wantConfig: PortConfig{
				ProxyProtocol:  "http",
				TargetProtocol: "http",
				ProxyPort:      80,
				TargetPort:     8080,
				IsRedirect:     false,
			},
			wantErr: false,
		},
		{
			name:  "Long: Valid proxy without protocols",
			input: "443" + proxySeparator + "80",
			wantConfig: PortConfig{
				ProxyProtocol:  "https",
				TargetProtocol: "http",
				ProxyPort:      443,
				TargetPort:     80,
				IsRedirect:     false,
			},
			wantErr: false,
		},
		{
			name:  "Long: Valid redirect with URL",
			input: "80" + protocolSeparator + "http" + redirectSeparator + "https:" + protocolSeparator + "/example.com",
			wantConfig: PortConfig{
				ProxyProtocol:  "http",
				ProxyPort:      80,
				IsRedirect:     true,
				TargetProtocol: "",
				TargetPort:     0,
				RedirectURL: func() *url.URL {
					u, _ := url.Parse("https:" + protocolSeparator + "/example.com")
					return u
				}(),
			},
			wantErr: false,
		},
		{
			name:  "Long: Valid redirect with URL without proxy protocol",
			input: "443" + redirectSeparator + "https:" + protocolSeparator + "/example.com",
			wantConfig: PortConfig{
				ProxyProtocol:  "https",
				ProxyPort:      443,
				IsRedirect:     true,
				TargetProtocol: "",
				TargetPort:     0,
				RedirectURL: func() *url.URL {
					u, _ := url.Parse("https://example.com")
					return u
				}(),
			},
			wantErr: false,
		},

		{
			name:  "Long: Valid redirect with URL without proxy protocol and with target port",
			input: "443" + redirectSeparator + "https:" + protocolSeparator + "/example.com:80",
			wantConfig: PortConfig{
				ProxyProtocol:  "https",
				ProxyPort:      443,
				IsRedirect:     true,
				TargetProtocol: "",
				TargetPort:     0,
				RedirectURL: func() *url.URL {
					u, _ := url.Parse("https://example.com:80")
					return u
				}(),
			},
			wantErr: false,
		},
		{
			name:    "Long: Invalid format missing separator",
			input:   "443" + protocolSeparator + "https80" + protocolSeparator + "http",
			wantErr: true,
		},
		{
			name:    "Long: Invalid proxy port",
			input:   "invalid" + protocolSeparator + "https" + proxySeparator + "80" + protocolSeparator + "http",
			wantErr: true,
		},
		{
			name:    "Long: Invalid target port",
			input:   "443" + protocolSeparator + "https" + proxySeparator + "invalid" + protocolSeparator + "http",
			wantErr: true,
		},
		{
			name:    "Long: Invalid URL for redirect",
			input:   "443" + protocolSeparator + "https" + redirectSeparator + "invalid-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfig, err := NewPortLongLabel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPortLongLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !comparePortConfig(gotConfig, tt.wantConfig) {
				t.Errorf("NewPortLongLabel() = %+v, want %+v", gotConfig, tt.wantConfig)
			}
		})
	}
}

func compareShortLabelPortConfig(a, b PortConfig) bool {
	return a.ProxyProtocol == b.ProxyProtocol &&
		a.ProxyPort == b.ProxyPort
}

func comparePortConfig(a, b PortConfig) bool {
	return compareURLs(a.RedirectURL, b.RedirectURL) &&
		a.ProxyProtocol == b.ProxyProtocol &&
		a.TargetProtocol == b.TargetProtocol &&
		a.ProxyPort == b.ProxyPort &&
		a.TargetPort == b.TargetPort &&
		a.IsRedirect == b.IsRedirect
}

func compareURLs(a, b *url.URL) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.String() == b.String()
}
