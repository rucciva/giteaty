package caddyplugin

import (
	"fmt"
	"testing"

	"github.com/caddyserver/caddy"
	"github.com/stretchr/testify/assert"
)

func TestParseConfiguration(t *testing.T) {
	data := []struct {
		input string
		cfg   *config
		err   error
	}{
		{
			input: `gitea-auth https://gitea.io`,
			cfg: &config{
				giteaURL: "https://gitea.io",
				basePath: "/",
			},
		},

		{
			input: `gitea-auth /test https://gitea.io`,
			cfg: &config{
				giteaURL: "https://gitea.io",
				basePath: "/test",
			},
		},

		{
			input: `
			gitea-auth /test https://gitea.io {
				repo
			}
			`,
			cfg: &config{
				giteaURL: "https://gitea.io",
				basePath: "/test",
				repo: &repoConfig{
					path: "/{owner}/{repo}",
				},
			},
		},

		{
			input: `
			gitea-auth /test https://gitea.io {
				repo /{owner}/{repo}
			}
			`,
			cfg: &config{
				giteaURL: "https://gitea.io",
				basePath: "/test",
				repo: &repoConfig{
					path: "/{owner}/{repo}",
				},
			},
		},

		{
			input: `
			gitea-auth /test https://gitea.io {
				repo rucciva/test
			}
			`,
			cfg: &config{
				giteaURL: "https://gitea.io",
				basePath: "/test",
				repo: &repoConfig{
					static: true,
					path:   "rucciva/test",
				},
			},
		},

		{
			input: `
			gitea-auth /test https://gitea.io {
				repo /{owner}/{repo} {
					orgFailover	
				}
			}
			`,
			cfg: &config{
				giteaURL: "https://gitea.io",
				basePath: "/test",
				repo: &repoConfig{
					path:        "/{owner}/{repo}",
					orgFailover: true,
				},
			},
		},

		{
			input: `
			gitea-auth /test https://gitea.io {
				repo rucciva/test {
					orgFailover	
					matchPermission
				}
			}
			`,
			cfg: &config{
				giteaURL: "https://gitea.io",
				basePath: "/test",
				repo: &repoConfig{
					static:          true,
					path:            "rucciva/test",
					orgFailover:     true,
					matchPermission: true,
				},
			},
		},

		{
			input: `
			gitea-auth /test  https://gitea.io/ {
				org 
			}`,
			cfg: &config{
				giteaURL: "https://gitea.io",

				basePath: "/test",
				org: &orgConfig{
					path: "/{org}",
					teams: map[string]bool{
						"owners": true,
					},
				},
			},
		},

		{
			input: `
			gitea-auth /test  https://gitea.io/ {
				org /my/{org}.js
			}`,
			cfg: &config{
				giteaURL: "https://gitea.io",

				basePath: "/test",
				org: &orgConfig{
					path: "/my/{org}.js",
					teams: map[string]bool{
						"owners": true,
					},
				},
			},
		},

		{
			input: `
			gitea-auth /test https://gitea.io {
				org rucciva
			}
			`,
			cfg: &config{
				giteaURL: "https://gitea.io",
				basePath: "/test",
				org: &orgConfig{
					static: true,
					path:   "rucciva",
					teams: map[string]bool{
						"owners": true,
					},
				},
			},
		},

		{
			input: `
			gitea-auth /test  https://gitea.io/ {
				org /my/{org}.js {
					teams tester developer
				}
			}`,
			cfg: &config{
				giteaURL: "https://gitea.io",

				basePath: "/test",
				org: &orgConfig{
					path: "/my/{org}.js",
					teams: map[string]bool{
						"tester":    true,
						"developer": true,
					},
				},
			},
		},
		{
			input: `
			gitea-auth /test  https://gitea.io/ {
				allowInsecure

				repo /re/{owner}/{repo}.js {
					orgFailover
					matchPermission
				}
				org /my/{org}.js {
					teams tester developer
				}
			}`,
			cfg: &config{
				giteaURL:           "https://gitea.io",
				giteaAllowInsecure: true,

				basePath: "/test",
				repo: &repoConfig{
					path:            "/re/{owner}/{repo}.js",
					matchPermission: true,
					orgFailover:     true,
				},
				org: &orgConfig{
					path: "/my/{org}.js",
					teams: map[string]bool{
						"tester":    true,
						"developer": true,
					},
				},
			},
		},
		{
			input: `
			gitea-auth /test  https://gitea.io/ {
				allowInsecure

				setBasicAuth somepassword
				repo rucciva/repo {
					orgFailover
					matchPermission
				}
				org myorg {
					teams tester developer
				}
			}`,
			cfg: &config{
				giteaURL:           "https://gitea.io",
				giteaAllowInsecure: true,
				setBasicAuth:       func(s string) *string { return &s }("somepassword"),
				basePath:           "/test",
				repo: &repoConfig{
					static:          true,
					path:            "rucciva/repo",
					matchPermission: true,
					orgFailover:     true,
				},
				org: &orgConfig{
					static: true,
					path:   "myorg",
					teams: map[string]bool{
						"tester":    true,
						"developer": true,
					},
				},
			},
		},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("SubTest#%d", i), func(t *testing.T) {
			c := caddy.NewTestController("http", d.input)
			r, err := parseConfiguration(c)
			assert.Equal(t, d.err, err)
			assert.Equal(t, d.cfg, r)
		})
	}
}
