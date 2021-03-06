package caddyhandler

import "net/http"

func (drt *Directive) assertRepoOrOrgMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, name := drt.getOwnerRepo(r)
		err := drt.assertRepo(r, owner, name)
		if err != nil {
			err = drt.assertOrg(r, drt.getOrg(r))
		}
		if err != nil {
			setReturn(r.Context(), handlerReturn{i: 403, err: err})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (drt *Directive) assertRepoAndOrgMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, name := drt.getOwnerRepo(r)
		err := drt.assertRepo(r, owner, name)
		if err != nil {
			setReturn(r.Context(), handlerReturn{i: 403, err: err})
			return
		}

		err = drt.assertOrg(r, drt.getOrg(r))
		if err != nil {
			setReturn(r.Context(), handlerReturn{i: 403, err: err})
			return
		}

		next.ServeHTTP(w, r)
	})
}
