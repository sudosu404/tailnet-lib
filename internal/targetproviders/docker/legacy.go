// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package docker

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/almeidapaulopt/tsdproxy/internal/model"
)

func (c *container) getLegacyPort() (model.PortConfig, error) {
	cPort := c.container.Config.Labels[LabelContainerPort]
	if cPort == "" {
		cPort = c.getIntenalPort()
	}

	cProtocol, hasProtocol := c.container.Config.Labels[LabelScheme]
	if !hasProtocol {
		cProtocol = "http"
	}

	port, err := model.NewPortLongLabel("443/https:" + cPort + "/" + cProtocol)
	if err != nil {
		return port, err
	}

	port.TLSValidate = c.getLabelBool(LabelTLSValidate, model.DefaultTLSValidate)
	port.Tailscale.Funnel = c.getLabelBool(LabelFunnel, model.DefaultTailscaleFunnel)

	return port, nil
}

// getIntenalPort method returns the container internal port
func (c *container) getIntenalPort() string {
	// If Label is defined, get the container port
	if customContainerPort, ok := c.container.Config.Labels[LabelContainerPort]; ok {
		return customContainerPort
	}

	for p := range c.container.NetworkSettings.Ports {
		return p.Port()
	}
	// in network_mode=host
	for p := range c.container.HostConfig.PortBindings {
		return p.Port()
	}

	return ""
}

// getExposedPort method returns the container port
func (c *container) getExposedPort(internalPort string) string {
	for p, b := range c.container.HostConfig.PortBindings {
		if p.Port() == internalPort {
			return b[0].HostPort
		}
	}

	// return the first exposed port
	for _, bindings := range c.container.HostConfig.PortBindings {
		if len(bindings) > 0 {
			return bindings[0].HostPort
		}
	}

	return ""
}

func (c *container) getImagePort() string {
	for p := range c.image.Config.ExposedPorts {
		return p.Port()
	}
	return ""
}

// getProxyHostname method returns the proxy URL from the container label.
func (c *container) getProxyHostname() (string, error) {
	// Set custom proxy URL if present the Label in the container
	if customName, ok := c.container.Config.Labels[LabelName]; ok {
		// validate url
		if _, err := url.Parse("https://" + customName); err != nil {
			return "", err
		}
		return customName, nil
	}

	return c.getName(), nil
}

// getTargetURL method returns the container target URL
func (c *container) getTargetURL(hostname string, iPort *url.URL) (*url.URL, error) {
	internalPort := iPort.Port()
	if internalPort == "" {
		internalPort = c.getIntenalPort()
	}
	exposedPort := c.getExposedPort(internalPort)
	imagePort := c.getImagePort()

	if exposedPort == "" && internalPort == "" && imagePort == "" {
		return nil, ErrNoPortFoundInContainer
	}

	// return localhost if container same as host to serve the dashboard
	if osname, err := os.Hostname(); err == nil && strings.HasPrefix(c.container.ID, osname) {
		return url.Parse("http://127.0.0.1:" + internalPort)
	}

	// set autodetect
	if c.autodetect {
		// repeat auto detect in case the container is not ready
		for try := range autoDetectTries {
			c.log.Info().Int("try", try).Msg("Trying to auto detect target URL")
			if port, err := c.tryConnectContainer(hostname, internalPort, exposedPort, imagePort); err == nil {
				return port, nil
			}
			// wait to container get ready in case of startup
			time.Sleep(autoDetectSleep)
		}
	}

	// auto detect failed or was disabled
	port := exposedPort
	if port == "" {
		port = internalPort
	}

	return url.Parse(iPort.Scheme + "://" + c.defaultTargetHostname + ":" + port)
}
