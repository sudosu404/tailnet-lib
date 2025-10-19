// SPDX-FileCopyrightText: 2025 Hector @sudosu404 <hector@email.gnx>
// SPDX-License-Identifier: AGPL3

package config

import (
	"fmt"
	"net"
	"os"

	"github.com/creasty/defaults"
)

const (
	DockerDefaultName            = "local"
	TailscaleDefaultProviderName = "default"
)

// generateDefaultProviders method Generate the config from environment variables
// used in 0.x.x versions
func (c *config) generateDefaultProviders() {
	// Legacy Hostname from DOCKER_HOST from environment
	//
	c.generateDockerConfig()

	c.generateTailscaleConfig()
}

// generateDockerConfig method generate the Docker Config provider from environment variables
func (c *config) generateDockerConfig() {
	// Legacy Hostname from DOCKER_HOST from environment
	//
	docker := new(DockerTargetProviderConfig)
	// set DockerConfig defaults
	if err := defaults.Set(docker); err != nil {
		fmt.Printf("Error loading defaults: %v", err)
	}
	if os.Getenv("DOCKER_HOST") != "" {
		docker.Host = os.Getenv("DOCKER_HOST")
	}

	if os.Getenv("TAILNET_HOSTNAME") != "" {
		docker.TargetHostname = os.Getenv("TAILNET_HOSTNAME")
	}

	// Check whether the hostname host.docker.internal can be resolved. This allows avoiding updates to the TargetHostname field in the configuration file.
	ip, err := net.LookupIP("host.docker.internal")
	if err == nil || len(ip) > 0 {
		docker.TargetHostname = "host.docker.internal"
	}

	c.Docker[DockerDefaultName] = docker
}

// generateTailscaleConfig method  generate the Tailscale Config provider from environment variables
func (c *config) generateTailscaleConfig() {
	ts := new(TailscaleServerConfig)
	// set TailscaleConfig defaults
	if err := defaults.Set(ts); err != nil {
		fmt.Printf("Error loading defaults: %v", err)
	}

	authKeyFile := os.Getenv("TAILNET_AUTHKEYFILE")
	authKey := os.Getenv("TAILNET_AUTHKEY")
	controlURL := os.Getenv("TAILNET_CONTROLURL")
	dataDir := os.Getenv("TAILNET_DATADIR")

	if authKeyFile != "" {
		var err error
		authKey, err = c.getAuthKeyFromFile(authKeyFile)
		if err != nil {
			fmt.Printf("Error loading auth key from file: %v", err)
		}
	}

	if authKey != "" {
		ts.AuthKey = authKey
	}
	if authKeyFile != "" {
		ts.AuthKeyFile = authKeyFile
	}

	if controlURL != "" {
		ts.ControlURL = controlURL
	}
	if dataDir != "" {
		c.Tailscale.DataDir = dataDir
	}

	c.Tailscale.Providers[TailscaleDefaultProviderName] = ts

	if c.DefaultProxyProvider == "" {
		c.DefaultProxyProvider = TailscaleDefaultProviderName
	}
}
