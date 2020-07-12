package caddyhandler

import (
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/go-chi/chi"
)

type orgConfig struct {
	nameStatic bool
	name       string
	teams      map[string]bool
}

func (drt *Directive) assertOrgMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := drt.assertOrg(r, drt.getOrg(r))
		if err != nil {
			setReturn(r.Context(), handlerReturn{i: 403, err: err})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (drt *Directive) getOrg(r *http.Request) (name string) {
	name = drt.org.name
	if !drt.org.nameStatic {
		name = chi.URLParam(r, name)
	}
	return
}

func (drt *Directive) assertOrg(req *http.Request, orgname string) (err error) {
	if len(drt.org.teams) > 0 {
		return drt.assertOrgTeam(req, orgname)
	}

	gcl := drt.newGiteaClient(req)
	orgs, err := gcl.ListMyOrgs(gitea.ListOrgsOptions{})
	if err != nil {
		return errUnauthorized
	}
	for _, org := range orgs {
		if strings.EqualFold(orgname, org.UserName) {
			return
		}
	}
	return errUnauthorized
}

func (drt *Directive) assertOrgTeam(req *http.Request, orgname string) (err error) {
	gcl := drt.newGiteaClient(req)
	teams, err := gcl.ListMyTeams(&gitea.ListTeamsOptions{})
	if err != nil {
		return errUnauthorized
	}
	for _, team := range teams {
		if !strings.EqualFold(team.Organization.UserName, orgname) {
			continue
		}
		if _, ok := drt.org.teams[strings.ToLower(team.Name)]; ok {
			return
		}
	}
	return errUnauthorized
}
