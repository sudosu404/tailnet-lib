// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package docker

import (
	"fmt"
	"net"
	"net/url"
)

// tryConnectContainer method tries to connect to the container
func (c *container) tryConnectContainer(hostname, internalPort, exposedPort, imagePort string) (*url.URL, error) {
	// test connection with the container using docker networking
	// try connecting to internal ip and internal port
	if internalPort != "" {
		port, err := c.tryInternalPort(hostname, internalPort)
		if err == nil {
			return port, nil
		}
		c.log.Debug().Err(err).Msg("Error connecting to internal port")
	}

	// try connecting to internal gateway and exposed port
	if exposedPort != "" {
		port, err := c.tryExposedPort(hostname, exposedPort)
		if err == nil {
			return port, nil
		}
		c.log.Debug().Err(err).Msg("Error connecting to exposed port")
	}

	if imagePort != "" {
		port, err := c.tryInternalPort(hostname, imagePort)
		if err == nil {
			return port, nil
		}
		port, err = c.tryExposedPort(hostname, imagePort)
		if err == nil {
			return port, nil
		}

		c.log.Debug().Err(err).Msg("Error to connect using image port")
	}

	return nil, &NoValidTargetFoundError{containerName: c.container.Name}
}

// tryInternalPort method tries to connect to the container internal ip and internal port
func (c *container) tryInternalPort(hostname, port string) (*url.URL, error) {
	c.log.Debug().Str("hostname", hostname).Str("port", port).Msg("trying to connect to internal port")
	for _, network := range c.container.NetworkSettings.Networks {
		if network.IPAddress == "" {
			continue
		}
		// try connecting to container IP and internal port
		if err := c.dial(network.IPAddress, port); err == nil {
			c.log.Info().Str("address", network.IPAddress).
				Str("port", port).Msg("Successfully connected using internal ip and internal port")
			return url.Parse(c.scheme + "://" + network.IPAddress + ":" + port)
		}
		c.log.Debug().Str("address", network.IPAddress).
			Str("port", port).Msg("Failed to connect")
	}
	// if the container is running in host mode,
	// try connecting to defaultBridgeAddress of the host and internal port.
	if c.container.HostConfig.NetworkMode == "host" && c.defaultBridgeAddress != "" {
		if err := c.dial(c.defaultBridgeAddress, port); err == nil {
			c.log.Info().Str("address", c.defaultBridgeAddress).Str("port", port).Msg("Successfully connected using defaultBridgeAddress and internal port")
			return url.Parse(c.scheme + "://" + c.defaultBridgeAddress + ":" + port)
		}

		c.log.Debug().Str("address", c.defaultBridgeAddress).Str("port", port).Msg("Failed to connect")
	}

	return nil, ErrNoValidTargetFoundForInternalPorts
}

// tryExposedPort method tries to connect to the container internal ip and exposed port
func (c *container) tryExposedPort(hostname, port string) (*url.URL, error) {
	for _, network := range c.container.NetworkSettings.Networks {
		if err := c.dial(network.Gateway, port); err == nil {
			c.log.Info().Str("address", network.Gateway).Str("port", port).Msg("Successfully connected using docker network gateway and exposed port")
			return url.Parse(c.scheme + "://" + network.Gateway + ":" + port)
		}

		c.log.Debug().Str("address", network.Gateway).Str("port", port).Msg("Failed to connect using docker network gateway and exposed port")
	}

	// try connecting to configured host and exposed port
	if err := c.dial(hostname, port); err == nil {
		c.log.Info().Str("address", hostname).Str("port", port).Msg("Successfully connected using configured host and exposed port")
		return url.Parse(c.scheme + "://" + hostname + ":" + port)
	}

	c.log.Debug().Str("address", hostname).Str("port", port).Msg("Failed to connect")
	return nil, ErrNoValidTargetFoundForExposedPorts
}

// dial method tries to connect to a host and port
func (c *container) dial(host, port string) error {
	address := host + ":" + port
	conn, err := net.DialTimeout("tcp", address, dialTimeout)
	if err != nil {
		return fmt.Errorf("error dialing %s: %w", address, err)
	}
	conn.Close()

	return nil
}
