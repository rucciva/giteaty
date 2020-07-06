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
	next       httpserver.Handler
	directives []*directive

	router http.Handler
}

func newHandler(next httpserver.Handler, drts []*directive) *handler {
	h := &handler{next: next, directives: drts}
	h.initRouter()
	return h
}

func (h *handler) initRouter() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i, err := h.next.ServeHTTP(w, r)
		setReturn(r.Context(), i, err)
	})
	router := chi.NewRouter()
	router.NotFound(http.HandlerFunc(next))

	for _, drt := range h.directives {
		router.Route(drt.basePath, func(r chi.Router) {

			if drt.repo == nil && drt.org == nil || drt.setBasicAuth != nil {
				r.Use(drt.assertUserMiddleware)
			}
			if drt.repo == nil && drt.org == nil {
				r.Handle("/*", next)
			}

			if repo := drt.repo; repo != nil {
				switch repo.static {
				case true:
					r.Use(drt.assertStaticRepoMiddleware)
					r.Handle("/*", next)
				case false:
					r.Handle(repo.path, drt.assertRepoMiddleware(next))
				}
			}

			if org := drt.org; org != nil {
				switch org.static {
				case true:
					r.Use(drt.assertStaticOrgTeamMiddleware)
					r.Handle("/*", next)
				case false:
					r.Handle(org.path, drt.assertOrgTeamMiddleware(next))
				}
			}

		})
	}

	h.router = router
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (i int, err error) {
	ret := &handlerReturn{i: 200}
	r = r.WithContext(context.WithValue(r.Context(), handlerReturnKey{}, ret))
	h.router.ServeHTTP(w, r)
	return ret.i, ret.err
}
