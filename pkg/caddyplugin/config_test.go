package caddyplugin

import (
	"testing"

	"github.com/caddyserver/caddy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfiguration(t *testing.T) {
	i := `
	gitea-auth /test  https://gitea.io/ {
		repo 
		org developers tester
		perm {

		}
	}
	`
	c := caddy.NewTestController("http", i)
	r, err := parseConfiguration(c)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "https://gitea.io", r.giteaURL)
	assert.Equal(t, "/test", r.basePath)
	assert.True(t, r.authzOrgTeams["developers"])
	assert.True(t, r.authzOrgTeams["tester"])
}
