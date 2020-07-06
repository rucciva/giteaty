package caddyplugin

import (
	"context"
	"errors"
	"net/http"

	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	"github.com/go-chi/chi"
)

type handlerReturnKey struct{}
type handlerReturn struct {
	i   int
	err error
}

var (
	errUnauthorized = errors.New("403 Forbidden")
)

func setReturn(ctx context.Context, i int, err error) {
	v := ctx.Value(handlerReturnKey{})
	ret, ok := v.(*handlerReturn)
	if !ok {
		return
	}
	ret.i, ret.err = i, err

}

type handler struct {
	next httpserver.Handler
	cfg  []*config

	router http.Handler
}

func newHandler(next httpserver.Handler, cfg []*config) *handler {
	h := &handler{next: next, cfg: cfg}
	h.initRouter()
	return h
}

func (h *handler) initRouter() {
	next := func(w http.ResponseWriter, r *http.Request) {
		i, err := h.next.ServeHTTP(w, r)
		setReturn(r.Context(), i, err)
	}
	router := chi.NewRouter()
	router.NotFound(http.HandlerFunc(next))

	for _, cfg := range h.cfg {

		router.Route(cfg.basePath, func(r chi.Router) {
			if cfg.repo == nil && cfg.org == nil || cfg.setBasicAuth != nil {
				r.Use(cfg.assertUserMiddleware)
			}

			if cfg.repo == nil && cfg.org == nil {
				r.Handle("/*", http.HandlerFunc(next))
			}

			if repo := cfg.repo; repo != nil {
				switch repo.static {
				case true:
					r.Use(cfg.assertStaticRepoMiddleware)
				case false:
					r.Handle(repo.path, cfg.assertRepoMiddleware(http.HandlerFunc(next)))
				}
			}

			if org := cfg.org; org != nil {
				switch org.static {
				case true:
					r.Use(cfg.assertStaticOrgTeamMiddleware)
				case false:
					r.Handle(org.path, cfg.assertOrgTeamMiddleware(http.HandlerFunc(next)))
				}
			}

		})
	}

	h.router = router
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (i int, err error) {
	ret := &handlerReturn{}
	r = r.WithContext(context.WithValue(r.Context(), handlerReturnKey{}, ret))
	h.router.ServeHTTP(w, r)
	return ret.i, ret.err
}
