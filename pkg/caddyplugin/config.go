package caddyplugin

import (
	"fmt"
	"net/url"

	"github.com/caddyserver/caddy"
)

type config struct {
	giteaURL           string
	giteaAllowInsecure bool

	basePath      string
	authzRepo     bool
	authzOrg      bool
	authzOrgTeams map[string]bool
	authzPerm     bool
}

func parseConfiguration(c *caddy.Controller) (r *config, err error) {
	c.Next() // skip plugin name

	r = new(config)
	var u string
	switch args := c.RemainingArgs(); len(args) {
	case 1:
		r.basePath, u = "/", args[0]
	case 2:
		r.basePath, u = args[0], args[1]
	default:
		return r, fmt.Errorf("gitea base url is required")
	}

	gurl, err := url.Parse(u)
	if err != nil {
		return r, fmt.Errorf("invalid url %s: %v", c.Val(), err)
	}
	r.giteaURL = gurl.Scheme + "://" + gurl.Host

	r.authzOrgTeams = make(map[string]bool)
	r.authzOrgTeams["owners"] = true
	if err = parseBlock(c, r); err != nil {
		return
	}

	return r, nil
}

func parseBlock(c *caddy.Controller, r *config) (err error) {
	for c.NextBlock() {
		switch v := c.Val(); v {
		case "allowInsecure":
			r.giteaAllowInsecure = true

		case "repo":
			r.authzRepo = true

		case "org":
			r.authzOrg = true
			for _, team := range c.RemainingArgs() {
				r.authzOrgTeams[team] = true
			}
		case "perm":
			if len(c.RemainingArgs()) != 0 {
				return fmt.Errorf("perm only take block")
			}
			r.authzPerm = true
			if err = parsePermissionBlock(c, r); err != nil {
				return
			}
		}
	}
	return
}

func parsePermissionBlock(c *caddy.Controller, r *config) (err error) {
	return
}
