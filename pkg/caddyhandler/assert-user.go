package caddyhandler

import (
	"net/http"
)

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

func (drt *Directive) assertUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := drt.assertUser(r)
		if err != nil {
			setReturn(r.Context(), handlerReturn{403, err, false})
			return
		}
		next.ServeHTTP(w, r)
	})
}
