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

func (drt *Directive) assertOrgTeam(req *http.Request, orgname string) (err error) {
	gcl := drt.newGiteaClient(req)
	teams, err := gcl.ListMyTeams(&gitea.ListTeamsOptions{})
	if err != nil {
		return errUnauthorized
	}
	for _, team := range teams {
		if strings.EqualFold(team.Organization.UserName, orgname) {
			if _, ok := drt.org.teams[strings.ToLower(team.Name)]; ok {
				return
			}
		}
	}
	return errUnauthorized
}

func (drt *Directive) assertOrgTeamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := drt.org.name
		if !drt.org.nameStatic {
			name = chi.URLParam(r, name)
		}
		err := drt.assertOrgTeam(r, name)
		if err != nil {
			setReturn(r.Context(), handlerReturn{403, err, false})
			return
		}
		next.ServeHTTP(w, r)
	})
}
