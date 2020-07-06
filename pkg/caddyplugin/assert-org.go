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

func (h *handler) assertOrgTeam(req *http.Request, orgname string) (err error) {
	gcl := h.newGiteaClient(req)
	teams, err := gcl.ListMyTeams(&gitea.ListTeamsOptions{})
	if err != nil {
		return errUnauthorized
	}
	for _, team := range teams {
		if strings.EqualFold(team.Organization.UserName, orgname) {
			if _, ok := h.cfg.org.teams[strings.ToLower(team.Name)]; ok {
				return
			}
		}
	}
	return errUnauthorized
}

func (h *handler) assertOrgTeamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h.assertOrgTeam(r, chi.URLParam(r, "org"))
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *handler) assertStaticOrgTeamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h.assertOrgTeam(r, h.cfg.org.path)
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
