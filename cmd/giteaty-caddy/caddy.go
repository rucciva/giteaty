package giteaty

import (
	"github.com/caddyserver/caddy"
	"github.com/rucciva/giteaty/pkg/caddyhandler"
)

func init() {
	caddy.RegisterPlugin("giteaty", caddy.Plugin{
		ServerType: "http",
		Action:     caddyhandler.Setup,
	})
}
