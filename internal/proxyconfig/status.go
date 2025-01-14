// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT
package proxyconfig

type ProxyState int32

const (
	ProxyStateInitializing ProxyState = iota
	ProxyStateStarting
	ProxyStateAuthenticating
	ProxyStateRunning
	ProxyStateStopping
	ProxyStateStopped
	ProxyStateError
)

var proxyStateStrings = []string{
	"Initializing",
	"Starting",
	"Authenticating",
	"Running",
	"Stopping",
	"Stopped",
	"Error",
}

func (s *ProxyState) String() string {
	return proxyStateStrings[int(*s)]
}

func (s *ProxyState) Int32() int32 {
	return int32(*s)
}
