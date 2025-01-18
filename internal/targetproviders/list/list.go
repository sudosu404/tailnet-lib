// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package list

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"reflect"
	"sync"

	"github.com/almeidapaulopt/tsdproxy/internal/config"
	"github.com/almeidapaulopt/tsdproxy/internal/proxyconfig"
	"github.com/almeidapaulopt/tsdproxy/internal/targetproviders"

	"github.com/creasty/defaults"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
)

type (
	// Client struct implements TargetProvider
	Client struct {
		log           zerolog.Logger
		file          *config.File
		configProxies configProxiesList
		proxies       configProxiesList
		eventsChan    chan targetproviders.TargetEvent
		errChan       chan error
		name          string
		config        config.FilesTargetProviderConfig
		mtx           sync.Mutex
	}

	configProxiesList map[string]proxyConfig

	proxyConfig struct {
		Ports         map[string]port       `yaml:"ports"`
		URL           string                `validate:"required,uri" yaml:"url"`
		ProxyProvider string                `yaml:"proxyProvider"`
		Dashboard     proxyconfig.Dashboard `validate:"dive" yaml:"dashboard"`
		Tailscale     proxyconfig.Tailscale `yaml:"tailscale"`
		TLSValidate   bool                  `validate:"boolean" default:"true" yaml:"tlsValidate"`
	}

	port struct {
		TLSValidate bool                      `validate:"boolean" default:"true" yaml:"tlsValidate"`
		RedirectURL string                    `yaml:"redirectUrl,omitempty"`
		Tailscale   proxyconfig.TailscalePort `validate:"dive" yaml:"tailscale"`
	}
)

var _ targetproviders.TargetProvider = (*Client)(nil)

func (s *proxyConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	_ = defaults.Set(s)

	type plain proxyConfig
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	return nil
}

// New function returns a new Files TargetProvider
func New(log zerolog.Logger, name string, provider *config.FilesTargetProviderConfig) (*Client, error) {
	newlog := log.With().Str("file", name).Logger()

	proxiesList := configProxiesList{}

	file := config.NewFile(newlog, provider.Filename, proxiesList)
	err := file.Load()
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	c := &Client{
		file:          file,
		log:           newlog,
		name:          name,
		configProxies: proxiesList,
		proxies:       make(map[string]proxyConfig),
		eventsChan:    make(chan targetproviders.TargetEvent),
		errChan:       make(chan error),
	}

	// load default values
	err = defaults.Set(c)
	if err != nil {
		return nil, fmt.Errorf("error loading defaults: %w", err)
	}

	return c, nil
}

func (c *Client) WatchEvents(_ context.Context, eventsChan chan targetproviders.TargetEvent, errChan chan error) {
	c.log.Debug().Msg("Start WatchEvents")

	c.eventsChan = eventsChan
	c.errChan = errChan

	c.file.Watch()
	c.file.OnChange(c.onFileChange)

	// start initial proxies
	go func() {
		for k := range c.configProxies {
			eventsChan <- targetproviders.TargetEvent{
				ID:             k,
				TargetProvider: c,
				Action:         targetproviders.ActionStartProxy,
			}
		}
	}()
}

func (c *Client) GetDefaultProxyProviderName() string {
	return c.config.DefaultProxyProvider
}

func (c *Client) Close() {
	for name := range c.proxies {
		c.eventsChan <- targetproviders.TargetEvent{
			ID:             name,
			TargetProvider: c,
			Action:         targetproviders.ActionStopProxy,
		}
	}

	close(c.eventsChan)
	close(c.errChan)
}

func (c *Client) AddTarget(id string) (*proxyconfig.Config, error) {
	proxy, ok := c.configProxies[id]
	if !ok {
		return nil, fmt.Errorf("target %s not found", id)
	}

	pcfg, err := c.newProxyConfig(id, proxy)
	if err != nil {
		return nil, err
	}

	return pcfg, nil
}

func (c *Client) DeleteProxy(id string) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if _, ok := c.proxies[id]; !ok {
		return fmt.Errorf("target %s not found", id)
	}

	delete(c.proxies, id)

	return nil
}

// newProxyConfig method returns a new proxyconfig.Config
func (c *Client) newProxyConfig(name string, p proxyConfig) (*proxyconfig.Config, error) {
	targetURL, err := url.Parse(p.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing target URL: %w", err)
	}

	proxyProvider := c.config.DefaultProxyProvider
	if p.ProxyProvider != "" {
		proxyProvider = p.ProxyProvider
	}

	proxyAccessLog := proxyconfig.DefaultProxyAccessLog

	pcfg, err := proxyconfig.NewConfig()
	if err != nil {
		return nil, err
	}

	pcfg.TargetID = name
	pcfg.TargetURL = targetURL
	pcfg.Hostname = name
	pcfg.TargetProvider = c.name
	pcfg.Tailscale = p.Tailscale
	pcfg.ProxyProvider = proxyProvider
	pcfg.ProxyAccessLog = proxyAccessLog
	pcfg.Ports = c.getPorts(p.Ports)
	// TODO:
	// pcfg.TLSValidate = p.TLSValidate
	pcfg.Dashboard = p.Dashboard

	c.addTarget(p, name)

	return pcfg, nil
}

func (c *Client) onFileChange(e fsnotify.Event) {
	if !e.Op.Has(fsnotify.Write) {
		return
	}
	c.log.Info().Str("filename", e.Name).Msg("config changed, reloading")
	oldConfigProxies := maps.Clone(c.configProxies)

	// Delete all entries because it's not deleted when loading from file
	for k := range c.configProxies {
		delete(c.configProxies, k)
	}
	if err := c.file.Load(); err != nil {
		c.log.Error().Err(err).Msg("error loading config")
	}

	// delete proxies that don't exist in new config
	for name := range oldConfigProxies {
		if _, ok := c.configProxies[name]; !ok {
			c.eventsChan <- targetproviders.TargetEvent{
				ID:             name,
				TargetProvider: c,
				Action:         targetproviders.ActionStopProxy,
			}
		}
	}

	for name := range c.configProxies {
		// start new proxies
		if _, ok := oldConfigProxies[name]; !ok {
			c.eventsChan <- targetproviders.TargetEvent{
				ID:             name,
				TargetProvider: c,
				Action:         targetproviders.ActionStartProxy,
			}
			continue
		}
		// restart if the proxy configuration changed
		if reflect.DeepEqual(c.configProxies[name], oldConfigProxies[name]) {
			c.eventsChan <- targetproviders.TargetEvent{
				ID:             name,
				TargetProvider: c,
				Action:         targetproviders.ActionRestartProxy,
			}
		}
	}
}

// addTarget method add a target the proxies map
func (c *Client) addTarget(cfg proxyConfig, name string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.proxies[name] = cfg
}

func (c *Client) getPorts(l map[string]port) map[string]proxyconfig.PortConfig {
	ports := make(map[string]proxyconfig.PortConfig)
	for k, v := range l {
		port, err := proxyconfig.NewPort(k)
		if err != nil {
			c.log.Error().Err(err).Str("port", k).Msg("error creating port config")
		}
		ports[k] = port
		port.TLSValidate = v.TLSValidate
		port.Tailscale = v.Tailscale
	}
	return ports
}
