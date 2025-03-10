// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package docker

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/almeidapaulopt/tsdproxy/internal/model"
	"github.com/almeidapaulopt/tsdproxy/web"

	ctypes "github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog"
)

// container struct stores the data from the docker container.
type container struct {
	log                   zerolog.Logger
	container             ctypes.InspectResponse
	defaultTargetHostname string
	defaultBridgeAddress  string
	targetProviderName    string
	autodetect            bool
}

// newContainer function returns a new container.
func newContainer(logger zerolog.Logger, dcontainer ctypes.InspectResponse,
	targetproviderName string, defaultBridgeAddress string, defaultTargetHostname string, providerAutoDetect bool,
) *container {
	//
	newlog := logger.With().Str("container", dcontainer.Name).Logger()
	newlog.Trace().Msg("New Container")
	defer newlog.Trace().Msg("End New Container")

	c := &container{
		log:                   newlog,
		container:             dcontainer,
		defaultTargetHostname: defaultTargetHostname,
		defaultBridgeAddress:  defaultBridgeAddress,
		targetProviderName:    targetproviderName,
	}

	c.autodetect = c.getLabelBool(LabelAutoDetect, providerAutoDetect)

	return c
}

// newProxyConfig method returns a new proxyconfig.Config.
func (c *container) newProxyConfig() (*model.Config, error) {
	c.log.Trace().Msg("New ProxyConfig")
	defer c.log.Trace().Msg("End New ProxyConfig")

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
	c.log.Trace().Msg("getPorts")
	defer c.log.Trace().Msg("End getPorts")

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
			port, err = c.generateTargetFromFirstTarget(port)
			if err == nil {
				ports[k] = port
			} else {
				c.log.Error().Err(err).Str("port", k).Msg("error generating target")
			}
		}
	}

	return ports
}

func (c *container) generateTargetFromFirstTarget(port model.PortConfig) (model.PortConfig, error) {
	c.log.Trace().Msg("generateTargetFromFirstTarget")
	defer c.log.Trace().Msg("End generateTargetFromFirstTarget")

	// multiple targets not supported in this TargetProvider
	p := port.GetFirstTarget()

	targetURL, err := c.getTargetURL(p)
	if err != nil {
		return port, err
	}
	c.log.Debug().Str("port", port.String()).Str("target", targetURL.String()).Msg("target URL")

	port.ReplaceTarget(p, targetURL)

	return port, nil
}

// getTailscaleConfig method returns the tailscale configuration.
func (c *container) getTailscaleConfig() (*model.Tailscale, error) {
	c.log.Trace().Msg("getTailscaleConfig")
	defer c.log.Trace().Msg("End getTailscaleConfig")

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

// getTargetURL method returns the container target URL
func (c *container) getTargetURL(iPort *url.URL) (*url.URL, error) {
	c.log.Trace().Msg("getTargetURL")
	defer c.log.Trace().Msg("End getTargetURL")

	internalPort := iPort.Port()
	publishedPort := c.getPublishedPort(internalPort)

	if internalPort == "" && publishedPort == "" {
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
			if port, err := c.tryConnectContainer(iPort.Scheme, internalPort, publishedPort); err == nil {
				return port, nil
			}
			// wait to container get ready in case of startup
			time.Sleep(autoDetectSleep)
		}
	}

	if c.container.HostConfig.NetworkMode == "host" && c.defaultBridgeAddress != "" {
		return url.Parse(iPort.Scheme + "://" + c.defaultTargetHostname + ":" + internalPort)
	}

	// auto detect failed or disabled, use published port
	if publishedPort == "" {
		return nil, ErrNoPortFoundInContainer
	}

	return url.Parse(iPort.Scheme + "://" + c.defaultTargetHostname + ":" + publishedPort)
}

// getPublishedPort method returns the container port
func (c *container) getPublishedPort(internalPort string) string {
	c.log.Trace().Msg("getPublishedPort")
	defer c.log.Trace().Msg("End getPublishedPort")

	for p, b := range c.container.HostConfig.PortBindings {
		if p.Port() == internalPort {
			return b[0].HostPort
		}
	}

	return ""
}

// getProxyHostname method returns the proxy URL from the container label.
func (c *container) getProxyHostname() (string, error) {
	c.log.Trace().Msg("getProxyHostname")
	defer c.log.Trace().Msg("End getProxyHostname")

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
