package caddyplugin

import (
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/go-chi/chi"
)

type orgConfig struct {
	static bool
	path   string
	teams  map[string]bool
}

func (drt *directive) assertOrgTeam(req *http.Request, orgname string) (err error) {
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

func (drt *directive) assertOrgTeamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := drt.assertOrgTeam(r, chi.URLParam(r, "org"))
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (drt *directive) assertStaticOrgTeamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := drt.assertOrgTeam(r, drt.org.path)
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
