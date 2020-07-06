package caddyplugin

import (
	"net/http"
)

func (h *config) assertUser(req *http.Request) (err error) {
	gcl := h.newGiteaClient(req)
	user, err := gcl.GetMyUserInfo()
	if err != nil {
		return errUnauthorized
	}
	if h.setBasicAuth != nil {
		req.SetBasicAuth(user.UserName, *h.setBasicAuth)
	}
	return
}

func (h *config) assertUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h.assertUser(r)
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
