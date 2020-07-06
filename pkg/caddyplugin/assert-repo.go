package caddyplugin

import (
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/go-chi/chi"
)

type repoConfig struct {
	static          bool
	path            string
	orgFailover     bool
	matchPermission bool
}

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

	if !h.cfg.repo.matchPermission {
		return
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
		if err != nil && h.cfg.repo.orgFailover && h.cfg.org != nil {
			err = h.assertOrgTeam(r, chi.URLParam(r, "owner"))
		}
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *handler) assertStaticRepoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.Split(h.cfg.org.path, "/")
		err := h.assertRepo(r, p[0], p[1])
		if err != nil && h.cfg.repo.orgFailover && h.cfg.org != nil {
			err = h.assertOrgTeam(r, p[0])
		}
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
