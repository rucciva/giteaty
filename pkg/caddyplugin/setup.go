package caddyplugin

import (
	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("gitea-auth", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) (err error) {
	cfg, err := parseConfiguration(c)
	if err != nil {
		return err
	}

	s := httpserver.GetConfig(c)
	s.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return newHandler(next, cfg)
	})
	return
}
