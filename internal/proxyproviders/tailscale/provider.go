// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"path"
	"path/filepath"
	"strings"

	"github.com/almeidapaulopt/tsdproxy/internal/config"
	"github.com/almeidapaulopt/tsdproxy/internal/model"
	"github.com/almeidapaulopt/tsdproxy/internal/proxyproviders"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2/clientcredentials"
	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

type (
	// Client struct implements proxyprovider for tailscale
	Client struct {
		log zerolog.Logger

		Hostname     string
		AuthKey      string
		clientID     string
		clientSecret string
		controlURL   string
		datadir      string
		tags         string
	}

	oauth struct {
		Authkey string `yaml:"authkey"`
	}
)

var _ proxyproviders.Provider = (*Client)(nil)

func New(log zerolog.Logger, name string, provider *config.TailscaleServerConfig) (*Client, error) {
	datadir := filepath.Join(config.Config.Tailscale.DataDir, name)

	tailscale.I_Acknowledge_This_API_Is_Unstable = true

	return &Client{
		log:          log.With().Str("tailscale", name).Logger(),
		Hostname:     name,
		AuthKey:      strings.TrimSpace(provider.AuthKey),
		clientID:     strings.TrimSpace(provider.ClientID),
		clientSecret: strings.TrimSpace(provider.ClientSecret),
		tags:         strings.TrimSpace(provider.Tags),
		datadir:      datadir,
		controlURL:   provider.ControlURL,
	}, nil
}

// NewProxy method implements proxyprovider NewProxy method
func (c *Client) NewProxy(config *model.Config) (proxyproviders.ProxyInterface, error) {
	c.log.Debug().
		Str("hostname", config.Hostname).
		Msg("Setting up tailscale server")

	log := c.log.With().Str("Hostname", config.Hostname).Logger()

	datadir := path.Join(c.datadir, config.Hostname)
	authKey := c.getAuthkey(config, datadir)

	tserver := &tsnet.Server{
		Hostname:     config.Hostname,
		AuthKey:      authKey,
		Dir:          datadir,
		Ephemeral:    config.Tailscale.Ephemeral,
		RunWebClient: config.Tailscale.RunWebClient,
		UserLogf: func(format string, args ...any) {
			log.Info().Msgf(format, args...)
		},
		Logf: func(format string, args ...any) {
			log.Trace().Msgf(format, args...)
		},

		ControlURL: c.getControlURL(),
	}

	// if verbose is set, use the info log level
	if config.Tailscale.Verbose {
		tserver.Logf = func(format string, args ...any) {
			log.Info().Msgf(format, args...)
		}
	}

	return &Proxy{
		log:      log,
		config:   config,
		tsServer: tserver,
		events:   make(chan model.ProxyEvent),
	}, nil
}

// getControlURL method returns the control URL
func (c *Client) getControlURL() string {
	if c.controlURL == "" {
		return model.DefaultTailscaleControlURL
	}
	return c.controlURL
}

func (c *Client) getAuthkey(config *model.Config, path string) string {
	authKey := config.Tailscale.AuthKey

	if c.clientID != "" && c.clientSecret != "" {
		authKey = c.getOAuth(config, path)
	}

	if authKey == "" {
		authKey = c.AuthKey
	}
	return authKey
}

func (c *Client) getOAuth(cfg *model.Config, dir string) string {
	baseURL := "https://api.tailscale.com"

	data := new(oauth)

	file := config.NewConfigFile(c.log, path.Join(dir, "tsdproxy.yaml"), data)
	if err := file.Load(); err == nil {
		if data.Authkey != "" {
			return data.Authkey
		}
	}

	credentials := clientcredentials.Config{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		TokenURL:     baseURL + "/api/v2/oauth/token",
	}

	ctx := context.Background()
	tsClient := tailscale.NewClient("-", nil)
	tsClient.UserAgent = "tsdproxy"
	tsClient.HTTPClient = credentials.Client(ctx)
	tsClient.BaseURL = baseURL

	temptags := strings.Trim(strings.TrimSpace(cfg.Tailscale.Tags), "\"")
	if temptags == "" {
		temptags = strings.Trim(strings.TrimSpace(c.tags), "\"")
	}

	if temptags == "" {
		c.log.Error().Msg("must define tags to use OAuth")
		return ""
	}

	caps := tailscale.KeyCapabilities{
		Devices: tailscale.KeyDeviceCapabilities{
			Create: tailscale.KeyDeviceCreateCapabilities{
				Reusable:  false,
				Ephemeral: cfg.Tailscale.Ephemeral,
				Tags:      strings.Split(temptags, ","),
			},
		},
	}

	authkey, _, err := tsClient.CreateKey(ctx, caps)
	if err != nil {
		c.log.Error().Err(err).Msg("unable to get Oauth token")
		return ""
	}

	data.Authkey = authkey
	if err := file.Save(); err != nil {
		c.log.Error().Err(err).Msg("unable to save oauth file")
	}

	return authkey
}
