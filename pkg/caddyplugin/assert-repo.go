package caddyplugin

import (
	"net/http"

	"code.gitea.io/sdk/gitea"
	"github.com/go-chi/chi"
)

type repoConfig struct {
	ownerStatic bool
	owner       string
	nameStatic  bool
	name        string
}

func (drt *Directive) assertRepo(req *http.Request, owner string, reponame string) (err error) {
	gcli := drt.newGiteaClient(req)
	_, err = gcli.GetRepo(owner, reponame)
	if err != nil {
		return errUnauthorized
	}
	_, err = gcli.ListRepoBranches(owner, reponame, gitea.ListRepoBranchesOptions{ListOptions: gitea.ListOptions{PageSize: 1}})
	if err != nil {
		return errUnauthorized
	}

	return
}

func (drt *Directive) assertRepoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, name := drt.repo.owner, drt.repo.name
		if !drt.repo.ownerStatic {
			owner = chi.URLParam(r, owner)
		}
		if !drt.repo.nameStatic {
			name = chi.URLParam(r, name)
		}
		err := drt.assertRepo(r, owner, name)
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (drt *Directive) assertRepoOrOrgMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, name := drt.repo.owner, drt.repo.name
		if !drt.repo.ownerStatic {
			owner = chi.URLParam(r, owner)
		}
		if !drt.repo.nameStatic {
			name = chi.URLParam(r, name)
		}

		err := drt.assertRepo(r, owner, name)
		if err != nil {
			drt.assertOrgTeamMiddleware(next).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
