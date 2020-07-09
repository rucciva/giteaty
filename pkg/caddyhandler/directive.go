package caddyhandler

import (
	"net/http"
)

type authz string

const (
	authzNone       authz = "none"
	authzUsers      authz = "users"
	authzRepo       authz = "repo"
	authzOrg        authz = "org"
	authzRepoOrOrg  authz = "repoOrOrg"
	authzRepoAndOrg authz = "repoAndOrg"
	authzDeny       authz = "deny"
)

var authzs = map[authz]bool{
	authzNone:       true,
	authzUsers:      true,
	authzRepo:       true,
	authzOrg:        true,
	authzRepoOrOrg:  true,
	authzRepoAndOrg: true,
	authzDeny:       true,
}

type Directive struct {
	giteaURL string
	insecure bool

	paths        []string
	methods      []string
	setBasicAuth *string

	authz authz
	users map[string]bool
	repo  *repoConfig
	org   *orgConfig
}

func (drt *Directive) handler(next http.Handler) http.Handler {
	var m func(http.Handler) http.Handler

	userAsserted := false
	switch drt.authz {
	case authzNone:
		m = drt.assertUserMiddleware
		userAsserted = true

	case authzUsers:
		m = drt.assertUserMiddleware
		userAsserted = true

	case authzRepo:
		m = drt.assertRepoMiddleware

	case authzOrg:
		m = drt.assertOrgTeamMiddleware

	case authzRepoOrOrg:
		m = drt.assertRepoOrOrgMiddleware

	case authzRepoAndOrg:
		m = drt.assertRepoAndOrgMiddleware

	default:
		return drt.denyMiddleware(next)
	}

	if drt.setBasicAuth != nil && !userAsserted {
		next = drt.assertUserMiddleware(next)
	}
	return drt.assertLoginMiddleware(m(next))
}

func (drt *Directive) denyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setReturn(r.Context(), handlerReturn{i: 403, err: errUnauthorized})
	})
}

func (drt *Directive) assertLoginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		ret := getReturn(r.Context())
		if ret == nil || ret.auth {
			return
		}

		// just ask password so that basic auth form always be displayed
		ret.i = 401
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	})
}
