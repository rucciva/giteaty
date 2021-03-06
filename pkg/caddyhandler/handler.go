package caddyhandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	"github.com/go-chi/chi"
)

var (
	errUnauthorized = errors.New("403 Forbidden")
)

func init() {
	// http://www.webdav.org/specs/rfc4918.html#http.methods.for.distributed.authoring
	customMethods := []string{"PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK"}
	for _, m := range customMethods {
		chi.RegisterMethod(m)
	}
}

type handlerReturnKey struct{}
type handlerReturn struct {
	i   int
	err error

	// auth signal wheter user is authenticated+authorized or not
	next bool
}

func getReturn(ctx context.Context) (ret *handlerReturn) {
	v := ctx.Value(handlerReturnKey{})
	ret, _ = v.(*handlerReturn)
	return
}

func setReturn(ctx context.Context, ret handlerReturn) {
	r := getReturn(ctx)
	if r == nil {
		return
	}
	r.i, r.err, r.next = ret.i, ret.err, ret.next
}

type handler struct {
	next       httpserver.Handler
	directives []*Directive

	router http.Handler
}

func NewHandler(next httpserver.Handler, drts []*Directive) *handler {
	h := &handler{next: next, directives: drts}
	h.initRouter()
	return h
}

func (h *handler) initRouter() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i, err := h.next.ServeHTTP(w, r)
		setReturn(r.Context(), handlerReturn{i: i, err: err, next: true})
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
