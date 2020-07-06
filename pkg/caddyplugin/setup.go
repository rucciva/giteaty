package caddyplugin

import (
	"fmt"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("gitea-auth", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func testConfig(c *config) (err error) {
	defer func() {
		if err1 := recover(); err1 != nil && err != nil {
			err = fmt.Errorf("invalid configuration: %s", err1)
		}
	}()
	newHandler(nil, c)
	return
}

func setup(c *caddy.Controller) (err error) {
	cfg, err := parseConfiguration(c)
	if err != nil {
		return err
	}
	if err = testConfig(cfg); err != nil {
		return
	}

	s := httpserver.GetConfig(c)
	s.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return newHandler(next, cfg)
	})
	return
}
