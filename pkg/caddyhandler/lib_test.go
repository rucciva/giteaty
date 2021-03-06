package caddyhandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestChiAssumption(t *testing.T) {
	type path struct {
		z, m, a, b, c, d, e int
	}
	serve := func(w http.ResponseWriter, r *http.Request) path {
		p := path{}
		rtr := chi.NewRouter()
		rtr.Route("/basic/{something}", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					p.m++
					next.ServeHTTP(w, r)
				})
			})
			r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p.a++
			}))
			r.Handle("/{test}/{info}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p.b++
			}))
			r.Handle("/test{test}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p.c++
			}))
			r.Handle("/test*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p.d++
			}))
		})
		rtr.Method("MOVE", "/syalala", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p.e++
		}))
		rtr.NotFound(func(w http.ResponseWriter, r *http.Request) {
			p.z++
		})
		rtr.ServeHTTP(w, r)
		return p
	}

	var w *httptest.ResponseRecorder
	var r *http.Request

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/basic/lala", nil)
	assert.Equal(t, path{m: 1, a: 1}, serve(w, r))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/basic/lala/info", nil)
	assert.Equal(t, path{m: 1, a: 1}, serve(w, r))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/basic/lala/tes/info/", nil)
	assert.Equal(t, path{m: 1, a: 1}, serve(w, r))

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/basic/lala/tes/info", nil)
	assert.Equal(t, path{m: 1, b: 1}, serve(w, r))

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/basic/lala/testinfo", nil)
	assert.Equal(t, path{m: 1, c: 1}, serve(w, r))

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/basic/lala/test/info", nil)
	assert.Equal(t, path{m: 1, d: 1}, serve(w, r))

	w = httptest.NewRecorder()
	r = httptest.NewRequest("MOVE", "/syalala", nil)
	assert.Equal(t, path{e: 1}, serve(w, r))
	assert.Equal(t, 200, w.Code)

}
