package caddyhandler

import (
	"net/http"
)

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

	if drt.setBasicAuth != nil {
		req.SetBasicAuth(user.UserName, *drt.setBasicAuth)
	}
	return
}
