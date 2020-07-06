package caddyplugin

import (
	"net/http"
)

func (h *handler) assertUser(req *http.Request) (err error) {
	gcl := h.newGiteaClient(req)
	user, err := gcl.GetMyUserInfo()
	if err != nil {
		return errUnauthorized
	}
	if h.cfg.setBasicAuth != nil {
		req.SetBasicAuth(user.UserName, *h.cfg.setBasicAuth)
	}
	return
}

func (h *handler) assertUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h.assertUser(r)
		if err != nil {
			setReturn(r.Context(), 403, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
