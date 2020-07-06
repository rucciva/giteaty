package caddyplugin

import (
	"net/http"
)

func (drt *directive) assertUser(req *http.Request) (err error) {
	gcl := drt.newGiteaClient(req)
	user, err := gcl.GetMyUserInfo()
	if err != nil {
		return errUnauthorized
	}
	if drt.setBasicAuth != nil {
		req.SetBasicAuth(user.UserName, *drt.setBasicAuth)
	}
	return
}

func (drt *directive) assertUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := drt.assertUser(r)
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
