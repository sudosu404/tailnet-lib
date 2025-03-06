// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package docker

import (
	"fmt"
	"strings"

	"github.com/almeidapaulopt/tsdproxy/internal/model"
	"github.com/almeidapaulopt/tsdproxy/web"
	dcontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/rs/zerolog"
)

// container struct stores the data from the docker container.
type container struct {
	log                   zerolog.Logger
	container             dcontainer.InspectResponse
	defaultTargetHostname string
	defaultBridgeAddress  string
	targetProviderName    string
	image                 image.InspectResponse
	autodetect            bool
	scheme                string
}

// newContainer function returns a new container.
func newContainer(logger zerolog.Logger, dcontainer dcontainer.InspectResponse, imageInfo image.InspectResponse,
	targetproviderName string, defaultBridgeAddress string, defaultTargetHostname string, providerAutoDetect bool,
) *container {
	c := &container{
		log:                   logger.With().Str("container", dcontainer.Name).Logger(),
		container:             dcontainer,
		image:                 imageInfo,
		defaultTargetHostname: defaultTargetHostname,
		defaultBridgeAddress:  defaultBridgeAddress,
		targetProviderName:    targetproviderName,
	}

	c.autodetect = c.getLabelBool(LabelAutoDetect, providerAutoDetect)

	return c
}

// newProxyConfig method returns a new proxyconfig.Config.
func (c *container) newProxyConfig() (*model.Config, error) {
	// Get the proxy URL
	//
	hostname, err := c.getProxyHostname()
	if err != nil {
		return nil, fmt.Errorf("error parsing Hostname: %w", err)
	}

	// Get the Tailscale configuration
	tailscale, err := c.getTailscaleConfig()
	if err != nil {
		return nil, err
	}

	pcfg, err := model.NewConfig()
	if err != nil {
		return nil, err
	}

	pcfg.TargetID = c.container.ID
	pcfg.Hostname = hostname
	pcfg.TargetProvider = c.targetProviderName
	pcfg.Tailscale = *tailscale
	pcfg.ProxyProvider = c.getLabelString(LabelProxyProvider, model.DefaultProxyProvider)
	pcfg.ProxyAccessLog = c.getLabelBool(LabelContainerAccessLog, model.DefaultProxyAccessLog)
	pcfg.Dashboard.Visible = c.getLabelBool(LabelDashboardVisible, model.DefaultDashboardVisible)
	pcfg.Dashboard.Label = c.getLabelString(LabelDashboardLabel, pcfg.Hostname)

	pcfg.Dashboard.Icon = c.getLabelString(LabelDashboardIcon, "")
	if pcfg.Dashboard.Icon == "" {
		pcfg.Dashboard.Icon = web.GuessIcon(c.container.Config.Image)
	}

	pcfg.Ports = c.getPorts()

	// add port from legacy labels if no port configured
	if len(pcfg.Ports) == 0 {
		if legacyPort, err := c.getLegacyPort(); err == nil {
			pcfg.Ports["legacy"] = legacyPort
		}
	}

	return pcfg, nil
}

func (c *container) getPorts() model.PortConfigList {
	ports := make(model.PortConfigList)
	for k, v := range c.container.Config.Labels {
		if !strings.HasPrefix(k, LabelPort) {
			continue
		}

		parts := strings.Split(v, ",")

		port, err := model.NewPortLongLabel(parts[0])
		if err != nil {
			c.log.Error().Err(err).Str("port", k).Msg("error creating port config")
			continue
		}

		for _, v := range parts[1:] {
			v = strings.TrimSpace(v)
			switch v {
			case PortOptionNoTLSValidate:
				port.TLSValidate = false
			case PortOptionTailscaleFunnel:
				port.Tailscale.Funnel = true
			}
		}

		if !port.IsRedirect {
			// multiple targets not supported in this TargetProvider
			p := port.GetFirstTarget()

			targetURL, err := c.getTargetURL(c.defaultTargetHostname, p)
			if err != nil {
				c.log.Error().Err(err).Msg("error parsing target hostname")
				return ports
			}

			port.AddTarget(targetURL)
		}

		ports[k] = port
	}

	return ports
}

// getTailscaleConfig method returns the tailscale configuration.
func (c *container) getTailscaleConfig() (*model.Tailscale, error) {
	authKey := c.getLabelString(LabelAuthKey, "")

	authKey, err := c.getAuthKeyFromAuthFile(authKey)
	if err != nil {
		return nil, fmt.Errorf("error setting auth key from file : %w", err)
	}

	return &model.Tailscale{
		Ephemeral:    c.getLabelBool(LabelEphemeral, model.DefaultTailscaleEphemeral),
		RunWebClient: c.getLabelBool(LabelRunWebClient, model.DefaultTailscaleRunWebClient),
		Verbose:      c.getLabelBool(LabelTsnetVerbose, model.DefaultTailscaleVerbose),
		AuthKey:      authKey,
	}, nil
}

// getName method returns the name of the container
func (c *container) getName() string {
	return strings.TrimLeft(c.container.Name, "/")
}
