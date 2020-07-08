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
		for _, path := range drt.paths {
			if len(drt.methods) == 0 {
				router.Handle(path, drt.handler(next))
				continue
			}
			for _, method := range drt.methods {
				router.Method(method, path, drt.handler(next))
			}
		}
	}

	h.router = router
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (i int, err error) {
	ret := &handlerReturn{i: 200}
	r = r.WithContext(context.WithValue(r.Context(), handlerReturnKey{}, ret))
	h.router.ServeHTTP(w, r)
	return ret.i, ret.err
}
