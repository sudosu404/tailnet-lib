// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT
package models

type ProxyStatus int

const (
	ProxyStatusInitializing ProxyStatus = iota
	ProxyStatusStarting
	ProxyStatusAuthenticating
	ProxyStatusRunning
	ProxyStatusStopping
	ProxyStatusStopped
	ProxyStatusError
)

var proxyStatusStrings = []string{
	"Initializing",
	"Starting",
	"Authenticating",
	"Running",
	"Stopping",
	"Stopped",
	"Error",
}

func (s *ProxyStatus) String() string {
	return proxyStatusStrings[int(*s)]
}
