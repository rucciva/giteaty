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

func (drt *directive) assertRepo(req *http.Request, owner string, reponame string) (err error) {
	gcli := drt.newGiteaClient(req)
	repo, err := gcli.GetRepo(owner, reponame)
	if err != nil {
		return errUnauthorized
	}
	_, err = gcli.ListRepoBranches(owner, reponame, gitea.ListRepoBranchesOptions{ListOptions: gitea.ListOptions{PageSize: 1}})
	if err != nil {
		return errUnauthorized
	}

	if !drt.repo.matchPermission {
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

func (drt *directive) assertRepoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := drt.assertRepo(r, chi.URLParam(r, "owner"), chi.URLParam(r, "repo"))
		if err != nil && drt.repo.orgFailover && drt.org != nil {
			err = drt.assertOrgTeam(r, chi.URLParam(r, "owner"))
		}
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (drt *directive) assertStaticRepoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.Split(drt.org.path, "/")
		err := drt.assertRepo(r, p[0], p[1])
		if err != nil && drt.repo.orgFailover && drt.org != nil {
			err = drt.assertOrgTeam(r, p[0])
		}
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
