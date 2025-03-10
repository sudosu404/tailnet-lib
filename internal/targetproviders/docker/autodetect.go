// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package docker

import (
	"fmt"
	"net"
	"net/url"
)

// tryConnectContainer method tries to connect to the container
func (c *container) tryConnectContainer(scheme, internalPort, publishedPort string) (*url.URL, error) {
	hostname := c.container.Config.Hostname

	if internalPort != "" {
		// test connection with the container using docker networking
		// try connecting to internal ip and internal port
		port, err := c.tryInternalPort(scheme, hostname, internalPort)
		if err == nil {
			return port, nil
		}
		c.log.Debug().Err(err).Msg("Error connecting to internal port")
	}

	// try connecting to internal gateway and published port
	if publishedPort != "" {
		port, err := c.tryPublishedPort(scheme, publishedPort)
		if err == nil {
			return port, nil
		}
		c.log.Debug().Err(err).Msg("Error connecting to published port")
	}

	return nil, &NoValidTargetFoundError{containerName: c.container.Name}
}

// tryInternalPort method tries to connect to the container internal ip and internal port
func (c *container) tryInternalPort(scheme, hostname, port string) (*url.URL, error) {
	c.log.Debug().Str("hostname", hostname).Str("port", port).Msg("trying to connect to internal port")
	for _, network := range c.container.NetworkSettings.Networks {
		if network.IPAddress == "" {
			continue
		}
		// try connecting to container IP and internal port
		if err := c.dial(network.IPAddress, port); err == nil {
			c.log.Info().Str("address", network.IPAddress).
				Str("port", port).Msg("Successfully connected using internal ip and internal port")
			return url.Parse(scheme + "://" + network.IPAddress + ":" + port)
		}
		c.log.Debug().Str("address", network.IPAddress).
			Str("port", port).Msg("Failed to connect")
	}
	// if the container is running in host mode,
	// try connecting to defaultBridgeAddress of the host and internal port.
	if c.container.HostConfig.NetworkMode == "host" && c.defaultBridgeAddress != "" {
		if err := c.dial(c.defaultBridgeAddress, port); err == nil {
			c.log.Info().Str("address", c.defaultBridgeAddress).Str("port", port).Msg("Successfully connected using defaultBridgeAddress and internal port")
			return url.Parse(scheme + "://" + c.defaultBridgeAddress + ":" + port)
		}

		c.log.Debug().Str("address", c.defaultBridgeAddress).Str("port", port).Msg("Failed to connect")
	}

	return nil, ErrNoValidTargetFoundForInternalPorts
}

// tryPublishedPort method tries to connect to the container internal ip and published port
func (c *container) tryPublishedPort(scheme, port string) (*url.URL, error) {
	for _, network := range c.container.NetworkSettings.Networks {
		if err := c.dial(network.Gateway, port); err == nil {
			c.log.Info().Str("address", network.Gateway).Str("port", port).Msg("Successfully connected using docker network gateway and published port")
			return url.Parse(scheme + "://" + network.Gateway + ":" + port)
		}

		c.log.Debug().Str("address", network.Gateway).Str("port", port).Msg("Failed to connect using docker network gateway and published port")
	}

	return nil, ErrNoValidTargetFoundForPublishedPorts
}

// dial method tries to connect to a host and port
func (c *container) dial(host, port string) error {
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, dialTimeout)
	if err != nil {
		return fmt.Errorf("error dialing %s: %w", address, err)
	}
	conn.Close()

	return nil
}
