package caddyplugin

import (
	"fmt"
	"testing"

	"github.com/caddyserver/caddy"
	"github.com/stretchr/testify/assert"
)

func TestParseDirectives(t *testing.T) {
	data := []struct {
		input      string
		directives []*directive
	}{
		{
			input: `
				gitea-auth https://gitea.io
				gitea-auth /test https://gitea.io
			`,
			directives: []*directive{
				{
					giteaURL: "https://gitea.io",
					basePath: "/",
				},
				{
					giteaURL: "https://gitea.io",
					basePath: "/test",
				},
			},
		},

		{
			input: `gitea-auth /test https://gitea.io`,
			directives: []*directive{
				{
					giteaURL: "https://gitea.io",
					basePath: "/test",
				},
			},
		},

		{
			input: `
			gitea-auth /test https://gitea.io {
				repo
			}
			`,
			directives: []*directive{
				{
					giteaURL: "https://gitea.io",
					basePath: "/test",
					repo: &repoConfig{
						path: "/{owner}/{repo}",
					},
				},
			},
		},

		{
			input: `
			gitea-auth /test https://gitea.io {
				repo /{owner}/{repo}
			}
			`,
			directives: []*directive{
				{
					giteaURL: "https://gitea.io",
					basePath: "/test",
					repo: &repoConfig{
						path: "/{owner}/{repo}",
					},
				},
			},
		},

		{
			input: `
			gitea-auth /test https://gitea.io {
				repo rucciva/test
			}
			`,
			directives: []*directive{
				{
					giteaURL: "https://gitea.io",
					basePath: "/test",
					repo: &repoConfig{
						static: true,
						path:   "rucciva/test",
					},
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
			directives: []*directive{
				{
					giteaURL: "https://gitea.io",
					basePath: "/test",
					repo: &repoConfig{
						path:        "/{owner}/{repo}",
						orgFailover: true,
					},
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
			directives: []*directive{
				{
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
		},

		{
			input: `
			gitea-auth /test  https://gitea.io/ {
				org 
			}`,
			directives: []*directive{
				{
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
		},

		{
			input: `
			gitea-auth /test  https://gitea.io/ {
				org /my/{org}.js
			}`,
			directives: []*directive{
				{
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
		},

		{
			input: `
			gitea-auth /test https://gitea.io {
				org rucciva
			}
			`,
			directives: []*directive{
				{
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
		},

		{
			input: `
			gitea-auth /test  https://gitea.io/ {
				org /my/{org}.js {
					teams tester developer
				}
			}`,
			directives: []*directive{
				{
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
			directives: []*directive{
				{
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
			directives: []*directive{
				{
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
			}
			gitea-auth /dev  https://gitea.io/ {
				allowInsecure

				setBasicAuth anotherpassword
				repo /{owner}/{repo} {
					orgFailover
					matchPermission
				}
				org /{org} {
					teams tester developer
				}
			}
			`,
			directives: []*directive{
				{
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
				{
					giteaURL:           "https://gitea.io",
					giteaAllowInsecure: true,
					setBasicAuth:       func(s string) *string { return &s }("anotherpassword"),
					basePath:           "/dev",
					repo: &repoConfig{
						path:            "/{owner}/{repo}",
						matchPermission: true,
						orgFailover:     true,
					},
					org: &orgConfig{
						path: "/{org}",
						teams: map[string]bool{
							"tester":    true,
							"developer": true,
						},
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
