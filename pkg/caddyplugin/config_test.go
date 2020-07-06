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
		repo /{owner}/{repo}.js {
			orgFailover
		}
		org /{org}.js {
			teams tester developer
		}
	}
	
	`
	c := caddy.NewTestController("http", i)

	r, err := parseConfiguration(c)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "https://gitea.io", r.giteaURL)
	assert.Equal(t, "/test", r.basePath)
	assert.Equal(t, "/{owner}/{repo}.js", r.repo.path)
	assert.True(t, r.repo.orgFailover)
	assert.Equal(t, "/{org}.js", r.org.path)
	assert.False(t, r.org.teams["owners"])
	assert.True(t, r.org.teams["developer"])
	assert.True(t, r.org.teams["tester"])
	assert.False(t, r.org.teams["unknown"])
}
