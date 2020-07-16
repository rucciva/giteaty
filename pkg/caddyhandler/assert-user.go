package caddyhandler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

type userConfig struct {
	nameStatic bool
	name       string
}

// assertUserMiddleware should be the last one since it can modify basic auth in case of setBasicAauth
func (drt *Directive) assertUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := drt.assertUser(r)
		if err != nil {
			setReturn(r.Context(), handlerReturn{i: 403, err: err})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (drt *Directive) assertUser(req *http.Request) (err error) {
	gcl := drt.newGiteaClient(req)
	user, err := gcl.GetMyUserInfo()
	if err != nil {
		return errUnauthorized
	}
	if drt.authz == authzUsers && !drt.users[user.UserName] {
		return errUnauthorized
	}
	if drt.authz == authzUser && !strings.EqualFold(drt.getUser(req), user.UserName) {
		return errUnauthorized
	}

	if drt.setBasicAuth != nil {
		req.SetBasicAuth(user.UserName, *drt.setBasicAuth)
	}
	return
}

func (drt *Directive) getUser(r *http.Request) (name string) {
	if drt.user == nil {
		return
	}

	name = drt.user.name
	if !drt.user.nameStatic {
		name = chi.URLParam(r, name)
	}
	return
}
