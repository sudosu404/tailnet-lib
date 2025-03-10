// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package docker

import "github.com/almeidapaulopt/tsdproxy/internal/model"

func (c *container) getLegacyPort() (model.PortConfig, error) {
	cPort := c.container.Config.Labels[LabelContainerPort]
	if cPort == "" {
		cPort = c.getIntenalPortLegacy()
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

	port, err = c.generateTargetFromFirstTarget(port)
	if err != nil {
		return port, err
	}

	return port, nil
}

// getIntenalPortLegacy method returns the container internal port
func (c *container) getIntenalPortLegacy() string {
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
