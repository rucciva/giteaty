package caddyplugin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/caddyserver/caddy"
)

type directive struct {
	giteaURL           string
	giteaAllowInsecure bool

	basePath     string
	setBasicAuth *string
	repo         *repoConfig
	org          *orgConfig
}

func parseDirectives(c *caddy.Controller) (drts []*directive, err error) {
	for c.Next() {
		cfg := new(directive)

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
		drts = append(drts, cfg)
	}

	return drts, nil
}

func parseBlock(c *caddy.Controller, drt *directive) (err error) {
	prevSection := ""
	for c.NextBlock() {
		v := c.Val()
		switch v {
		case "allowInsecure":
			drt.giteaAllowInsecure = true

		case "setBasicAuth":
			if drt.setBasicAuth != nil {
				return fmt.Errorf("can only have one 'setBasicAuth' section")
			}
			args := c.RemainingArgs()
			if len(args) != 1 {
				return fmt.Errorf("setBasicAuth take 1 password arg")
			}
			drt.setBasicAuth = &args[0]

		case "repo":
			if drt.repo != nil {
				return fmt.Errorf("can only have one 'repo' section")
			}
			drt.repo = &repoConfig{
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
				drt.repo.static = true
				drt.repo.path = args[0]
				break
			}

			if strings.Count(args[0], "{owner}") != 1 || strings.Count(args[0], "{repo}") != 1 {
				return fmt.Errorf("path should have 1 '{owner}' and 1 '{repo}' element")
			}
			drt.repo.path = args[0]

		case "org":
			if drt.org != nil {
				return fmt.Errorf("can only have one 'org' section")
			}
			drt.org = &orgConfig{
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
				drt.org.static = true
				drt.org.path = args[0]
				break
			}

			if strings.Count(args[0], "{org}") != 1 {
				return fmt.Errorf("path should have 1 '{org}' element")
			}
			drt.org.path = args[0]

		case "{":
			switch prevSection {
			case "repo":
				err = parseRepoSubBlock(c, drt)
			case "org":
				err = parseOrgSubBlock(c, drt)
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

func parseRepoSubBlock(c *caddy.Controller, drt *directive) (err error) {
	for next := c.Next(); next && c.Val() != "}"; next = c.Next() {
		switch v := c.Val(); v {
		case "orgFailover":
			drt.repo.orgFailover = true
		case "matchPermission":
			drt.repo.matchPermission = true
		default:
			return fmt.Errorf("unknwon keyword '%s' in organization block", v)
		}
	}

	return
}

func parseOrgSubBlock(c *caddy.Controller, drt *directive) (err error) {
	delete(drt.org.teams, "owners")

	for next := c.Next(); next && c.Val() != "}"; next = c.Next() {
		switch v := c.Val(); v {
		case "teams":
			for _, arg := range c.RemainingArgs() {
				drt.org.teams[arg] = true
			}
		default:
			return fmt.Errorf("unknwon keyword '%s' in organization block", v)
		}
	}

	if len(drt.org.teams) == 0 {
		drt.org.teams["owners"] = true
	}
	return
}

func validateDirectives(drts []*directive) (err error) {
	defer func() {
		if err1 := recover(); err1 != nil && err != nil {
			err = fmt.Errorf("invalid configuration: %s", err1)
		}
	}()
	newHandler(nil, drts)
	return
}
