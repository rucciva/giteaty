package caddyhandler

import (
	"fmt"
	"net/url"
	"strings"

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

func NewDirectives(c *caddy.Controller) (drts []*Directive, err error) {
	drts, err = parseDirectives(c)
	if err != nil {
		return
	}
	if err = validateDirectives(drts); err != nil {
		return
	}
	return
}

func parseDirectives(c *caddy.Controller) (drts []*Directive, err error) {
	for c.Next() {
		drt := &Directive{
			authz: authzNone,
		}

		args := c.RemainingArgs()
		if len(args) != 1 {
			return nil, fmt.Errorf("can only have 1 gitea base url defined")
		}
		gurl, err := url.Parse(args[0])
		if err != nil {
			return nil, fmt.Errorf("invalid url %s: %v", c.Val(), err)
		}
		drt.giteaURL = gurl.Scheme + "://" + gurl.Host

		if err = parseBlock(c, drt); err != nil {
			return nil, err
		}
		drts = append(drts, drt)
	}

	return drts, nil
}

func parseBlock(c *caddy.Controller, drt *Directive) (err error) {
	azIsSet, prevSection := false, ""
	for c.NextBlock() {
		v := c.Val()
		switch v {
		case "insecure":
			if drt.insecure {
				return fmt.Errorf("can only have one 'insecure' section")
			}
			c.RemainingArgs()
			drt.insecure = true

		case "paths":
			drt.paths = append(drt.paths, c.RemainingArgs()...)

		case "methods":
			for _, arg := range c.RemainingArgs() {
				drt.methods = append(drt.methods, strings.ToUpper(arg))
			}

		case "noauth":
			if drt.noauth {
				return fmt.Errorf("can only have one 'noauth' section")
			}
			c.RemainingArgs()
			drt.noauth = true

		case "authz":
			if azIsSet {
				return fmt.Errorf("can only have one 'authz' section")
			}
			args := c.RemainingArgs()
			if len(args) != 1 {
				return fmt.Errorf("'authz' takes exactly 1 arg")
			}
			if !authzs[authz(args[0])] {
				return fmt.Errorf("unknown 'authz' %s", args[0])
			}
			drt.authz = authz(args[0])
			azIsSet = true

		case "realm":
			if drt.realm != "" {
				return fmt.Errorf("can only have one 'realm' section")
			}
			args := c.RemainingArgs()
			if len(args) != 1 {
				return fmt.Errorf("'realm' only take 1 arg")
			}
			drt.realm = args[0]

		case "setBasicAuth":
			if drt.setBasicAuth != nil {
				return fmt.Errorf("can only have one 'setBasicAuth' section")
			}
			args := c.RemainingArgs()
			if len(args) != 1 {
				return fmt.Errorf("setBasicAuth take 1 password arg")
			}
			drt.setBasicAuth = &args[0]

		case "users":
			if drt.users == nil {
				drt.users = make(map[string]bool)
			}
			for _, arg := range c.RemainingArgs() {
				drt.users[arg] = true
			}

		case "repo":
			if drt.repo != nil {
				return fmt.Errorf("can only have one 'repo' section")
			}
			drt.repo = &repoConfig{
				owner: "owner",
				name:  "repo",
			}
			args := c.RemainingArgs()
			if len(args) > 2 || len(args) == 1 {
				return fmt.Errorf("repo can only takes exactly 2 args or none")
			}
			if len(args) == 0 {
				break
			}
			drt.repo.owner, drt.repo.ownerStatic = parsePathParameterName(args[0])
			drt.repo.name, drt.repo.nameStatic = parsePathParameterName(args[1])

		case "org":
			if drt.org != nil {
				return fmt.Errorf("can only have one 'org' section")
			}
			drt.org = &orgConfig{
				name: "org",
			}
			args := c.RemainingArgs()
			if len(args) > 1 {
				return fmt.Errorf("org only takes max 1 args")
			}
			if len(args) == 0 {
				break
			}
			drt.org.name, drt.org.nameStatic = parsePathParameterName(args[0])

		case "{":
			switch prevSection {
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
	if len(drt.paths) == 0 {
		drt.paths = append(drt.paths, "/")
	}
	return
}

func parsePathParameterName(s string) (p string, static bool) {
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		return s[1 : len(s)-1], false
	}
	return s, true
}

func parseOrgSubBlock(c *caddy.Controller, drt *Directive) (err error) {
	for next := c.Next(); next && c.Val() != "}"; next = c.Next() {
		switch v := c.Val(); v {
		case "teams":
			if drt.org.teams == nil {
				drt.org.teams = make(map[string]bool)
			}
			for _, arg := range c.RemainingArgs() {
				drt.org.teams[strings.ToLower(arg)] = true
			}
		default:
			return fmt.Errorf("unknwon keyword '%s' in 'org' block", v)
		}
	}
	return
}

func validateDirectives(drts []*Directive) (err error) {
	defer func() {
		if err1 := recover(); err1 != nil && err != nil {
			err = fmt.Errorf("invalid configuration: %s", err1)
		}
	}()
	for _, drt := range drts {
		if drt.authz == authzRepo || drt.authz == authzRepoAndOrg || drt.authz == authzRepoOrOrg {
			if drt.repo == nil {
				return fmt.Errorf("`%s` require `repo` section", drt.authz)
			}
		}
		if drt.authz == authzOrg || drt.authz == authzRepoAndOrg || drt.authz == authzRepoOrOrg {
			if drt.org == nil {
				return fmt.Errorf("`%s` require `org` section", drt.authz)
			}
		}
	}
	NewHandler(nil, drts)
	return
}
