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

func parseConfiguration(c *caddy.Controller) (cfgs []*config, err error) {
	for c.Next() {
		cfg := new(config)

		var u string
		switch args := c.RemainingArgs(); len(args) {
		case 1:
			cfg.basePath, u = "/", args[0]
		case 2:
			cfg.basePath, u = args[0], args[1]
		default:
			return nil, fmt.Errorf("gitea base url is required")
		}
		gurl, err := url.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("invalid url %s: %v", c.Val(), err)
		}
		cfg.giteaURL = gurl.Scheme + "://" + gurl.Host

		if err = parseBlock(c, cfg); err != nil {
			return nil, err
		}
		cfgs = append(cfgs, cfg)
	}

	return cfgs, nil
}

func parseBlock(c *caddy.Controller, cfg *config) (err error) {
	prevSection := ""
	for c.NextBlock() {
		v := c.Val()
		switch v {
		case "allowInsecure":
			cfg.giteaAllowInsecure = true

		case "setBasicAuth":
			if cfg.setBasicAuth != nil {
				return fmt.Errorf("can only have one 'setBasicAuth' section")
			}
			args := c.RemainingArgs()
			if len(args) != 1 {
				return fmt.Errorf("setBasicAuth take 1 password arg")
			}
			cfg.setBasicAuth = &args[0]

		case "repo":
			if cfg.repo != nil {
				return fmt.Errorf("can only have one 'repo' section")
			}
			cfg.repo = &repoConfig{
				path: "/{owner}/{repo}",
			}
			args := c.RemainingArgs()
			if len(args) > 1 {
				return fmt.Errorf("repo only takes max 1 args")
			}

			if len(args) == 0 {
				break
			}

			if !strings.HasPrefix(args[0], "/") && strings.Count(args[0], "/") == 1 {
				cfg.repo.static = true
				cfg.repo.path = args[0]
				break
			}

			if strings.Count(args[0], "{owner}") != 1 || strings.Count(args[0], "{repo}") != 1 {
				return fmt.Errorf("path should have 1 '{owner}' and 1 '{repo}' element")
			}
			cfg.repo.path = args[0]

		case "org":
			if cfg.org != nil {
				return fmt.Errorf("can only have one 'org' section")
			}
			cfg.org = &orgConfig{
				path:  "/{org}",
				teams: map[string]bool{"owners": true},
			}
			args := c.RemainingArgs()
			if len(args) > 1 {
				return fmt.Errorf("org only takes max 1 args")
			}

			if len(args) == 0 {
				break
			}

			if !strings.HasPrefix(args[0], "/") && strings.Count(args[0], "/") == 0 {
				cfg.org.static = true
				cfg.org.path = args[0]
				break
			}

			if strings.Count(args[0], "{org}") != 1 {
				return fmt.Errorf("path should have 1 '{org}' element")
			}
			cfg.org.path = args[0]

		case "{":
			switch prevSection {
			case "repo":
				err = parseRepoSubBlock(c, cfg)
			case "org":
				err = parseOrgSubBlock(c, cfg)
			default:
				err = fmt.Errorf("'%s' is not a sub block", prevSection)
			}
			if err != nil {
				return
			}

		default:
			return fmt.Errorf("invalid '%s' key", v)
		}
		prevSection = v
	}
	return
}

func parseRepoSubBlock(c *caddy.Controller, cfg *config) (err error) {
	for next := c.Next(); next && c.Val() != "}"; next = c.Next() {
		switch v := c.Val(); v {
		case "orgFailover":
			cfg.repo.orgFailover = true
		case "matchPermission":
			cfg.repo.matchPermission = true
		default:
			return fmt.Errorf("unknwon keyword '%s' in organization block", v)
		}
	}

	return
}

func parseOrgSubBlock(c *caddy.Controller, cfg *config) (err error) {
	delete(cfg.org.teams, "owners")

	for next := c.Next(); next && c.Val() != "}"; next = c.Next() {
		switch v := c.Val(); v {
		case "teams":
			for _, arg := range c.RemainingArgs() {
				cfg.org.teams[arg] = true
			}
		default:
			return fmt.Errorf("unknwon keyword '%s' in organization block", v)
		}
	}

	if len(cfg.org.teams) == 0 {
		cfg.org.teams["owners"] = true
	}
	return
}
