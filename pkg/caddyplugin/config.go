package caddyplugin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/caddyserver/caddy"
)

type config struct {
	giteaURL           string
	giteaAllowInsecure bool

	basePath     string
	setBasicAuth *string
	repo         *repoConfig
	org          *orgConfig
}

func parseConfiguration(c *caddy.Controller) (r *config, err error) {
	c.NextArg() // skip plugin name

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

	if err = parseBlock(c, r); err != nil {
		return
	}

	return r, nil
}

func parseBlock(c *caddy.Controller, r *config) (err error) {
	prevKey := ""
	for c.NextBlock() {
		v := c.Val()
		switch v {
		case "allowInsecure":
			r.giteaAllowInsecure = true

		case "setBasicAuth":
			args := c.RemainingArgs()
			if len(args) != 1 {
				return fmt.Errorf("setBasicAuth take 1 password arg")
			}
			r.setBasicAuth = &args[0]

		case "repo":
			r.repo = &repoConfig{
				path: "/{owner}/{repo}",
			}
			args := c.RemainingArgs()
			if len(args) > 1 {
				return fmt.Errorf("repo only takes max 1 args")
			}
			if len(args) == 0 {
				continue
			}
			if strings.Count(args[0], "{owner}") != 1 || strings.Count(args[0], "{repo}") != 1 {
				return fmt.Errorf("path should have 1 '{owner}' and 1 '{repo}' element")
			}
			r.repo.path = args[0]

		case "org":
			r.org = &orgConfig{
				path:  "/{org}",
				teams: map[string]bool{"owners": true},
			}
			args := c.RemainingArgs()
			if len(args) > 1 {
				return fmt.Errorf("org only takes max 1 args")
			}
			if len(args) == 0 {
				continue
			}
			if strings.Count(args[0], "{org}") != 1 {
				return fmt.Errorf("path should have 1 '{org}' element")
			}
			r.org.path = args[0]

		case "{":
			switch prevKey {
			case "repo":
				err = parseRepoSubBlock(c, r)
			case "org":
				err = parseOrgSubBlock(c, r)
			default:
				err = fmt.Errorf("'%s' is not a sub block", prevKey)
			}
			if err != nil {
				return
			}

		default:
			return fmt.Errorf("invalid '%s' key", v)
		}
		prevKey = v
	}
	return
}

func parseRepoSubBlock(c *caddy.Controller, r *config) (err error) {
	for next := c.Next(); next && c.Val() != "}"; next = c.Next() {
		switch v := c.Val(); v {
		case "orgFailover":
			r.repo.orgFailover = true
		case "matchPermission":
			r.repo.matchPermission = true
		default:
			return fmt.Errorf("unknwon keyword '%s' in organization block", v)
		}
	}

	return
}

func parseOrgSubBlock(c *caddy.Controller, r *config) (err error) {
	delete(r.org.teams, "owners")

	for next := c.Next(); next && c.Val() != "}"; next = c.Next() {
		switch v := c.Val(); v {
		case "teams":
			for _, arg := range c.RemainingArgs() {
				r.org.teams[arg] = true
			}
		default:
			return fmt.Errorf("unknwon keyword '%s' in organization block", v)
		}
	}

	if len(r.org.teams) == 0 {
		r.org.teams["owners"] = true
	}
	return
}
