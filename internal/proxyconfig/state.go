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

func ProxyStateString(s ProxyState) string {
	return proxyStateStrings[s]
}
