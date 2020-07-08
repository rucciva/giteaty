package caddyhandler

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/caddyserver/caddy"
	"github.com/stretchr/testify/assert"
)

func TestParseDirectives(t *testing.T) {
	data := []struct {
		input      string
		directives []*Directive
	}{
		{
			input: `
			giteaty https://gitea.io/
			`,
			directives: []*Directive{
				{
					giteaURL: "https://gitea.io",
					paths:    []string{"/"},
					authz:    authzNone,
				},
			},
		},
		{
			input: `
			giteaty https://gitea.io/
			giteaty https://gitea.io/ {
				insecure
				paths /test1 /test2
				methods gEt pOsT
				authz users
				
				users a b c
			}
			giteaty https://gitea.io/ {
				paths /test3/{user} /test4/{user}
				paths /test5/{user}/*
				methods gEt pOsT
				methods PUT patch
				authz repoAndOrg

				repo {user} test 
				org {user}
			}
			`,
			directives: []*Directive{
				{
					giteaURL: "https://gitea.io",
					paths:    []string{"/"},
					authz:    authzNone,
				},
				{
					giteaURL: "https://gitea.io",
					insecure: true,
					paths:    []string{"/test1", "/test2"},
					methods:  []string{http.MethodGet, http.MethodPost},
					authz:    authzUsers,
					users:    map[string]bool{"a": true, "b": true, "c": true},
				},
				{
					giteaURL: "https://gitea.io",
					paths:    []string{"/test3/{user}", "/test4/{user}", "/test5/{user}/*"},
					methods:  []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch},
					authz:    authzRepoAndOrg,
					repo: &repoConfig{
						owner: "user", name: "test", nameStatic: true,
					},
					org: &orgConfig{
						name:  "user",
						teams: map[string]bool{"owners": true},
					},
				},
			},
		},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("SubTest#%d", i), func(t *testing.T) {
			c := caddy.NewTestController("http", d.input)
			drt, err := parseDirectives(c)
			assert.NoError(t, err)
			assert.NoError(t, validateDirectives(drt))
			assert.ElementsMatch(t, d.directives, drt)
		})
	}
}
