package caddyplugin

import (
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/go-chi/chi"
)

func (h *handler) assertRepo(req *http.Request, owner string, reponame string) (err error) {
	gcli := h.newGiteaClient(req)
	repo, err := gcli.GetRepo(owner, reponame)
	if err != nil {
		return errUnauthorized
	}
	_, err = gcli.ListRepoBranches(owner, reponame, gitea.ListRepoBranchesOptions{ListOptions: gitea.ListOptions{PageSize: 1}})
	if err != nil {
		return errUnauthorized
	}
	switch strings.ToUpper(req.Method) {

	// Read repo code
	case http.MethodHead:
		return
	case http.MethodGet:
		return

	// Write repo code
	case http.MethodPost:
		fallthrough
	case http.MethodPatch:
		fallthrough
	case http.MethodPut:
		if repo.Permissions.Push {
			return
		}

	// Delete repo code
	case http.MethodDelete:
		if repo.Permissions.Admin {
			return
		}
	}
	return errUnauthorized
}

func (h *handler) assertRepoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h.assertRepo(r, chi.URLParam(r, "owner"), chi.URLParam(r, "repo"))
		if err != nil && h.cfg.authzOrg {
			err = h.assertOrgTeam(r, chi.URLParam(r, "owner"))
		}
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
