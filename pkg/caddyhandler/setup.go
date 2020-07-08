package caddyhandler

import (
	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

func Setup(c *caddy.Controller) (err error) {
	drts, err := NewDirectives(c)
	if err != nil {
		return
	}

	s := httpserver.GetConfig(c)
	s.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return NewHandler(next, drts)
	})
	return
}
