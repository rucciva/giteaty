package giteaty

import (
	"github.com/caddyserver/caddy"
	"github.com/rucciva/giteaty/pkg/caddyplugin"
)

func init() {
	caddy.RegisterPlugin("giteaty", caddy.Plugin{
		ServerType: "http",
		Action:     caddyplugin.Setup,
	})
}
