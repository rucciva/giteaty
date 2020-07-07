package caddyplugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testResponse struct {
	code    int
	headers http.Header
	body    interface{}
}

type testHandler struct {
	handler *handler

	reqGitea []*http.Request
	reqFwd   *http.Request
}

func tNewHandler(t *testing.T, conf string, ress []testResponse) (h *testHandler) {
	h = &testHandler{}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.reqGitea = append(h.reqGitea, r)

		res := ress[len(h.reqGitea)-1]
		for k, vs := range res.headers {
			for _, v := range vs {
				w.Header().Set(k, v)
			}
		}
		w.WriteHeader(res.code)
		_ = json.NewEncoder(w).Encode(res.body)
	}))

	i := fmt.Sprintf(conf, s.URL)
	c := caddy.NewTestController("http", i)
	require.NoError(t, setup(c), i)

	n := httpserver.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (i int, err error) {
		h.reqFwd = r
		return 201, nil
	})
	h.handler, _ = httpserver.GetConfig(c).Middleware()[0](n).(*handler)
	return
}

func TestHandlerAssertUser(t *testing.T) {
	user, pass := "user", "pass"
	conf := `gitea-auth %s`
	ress := []testResponse{
		{200, nil, &gitea.User{UserName: user}},
	}

	w := &httptest.ResponseRecorder{}
	r := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader(`{"message" : "hi"}`))
	r.SetBasicAuth(user, pass)
	h := tNewHandler(t, conf, ress)
	i, err := h.handler.ServeHTTP(w, r)

	assert.Equal(t, w.Code, 0, "should not write status code")
	assert.Empty(t, w.Header(), "should not write header")
	assert.Nil(t, w.Body, "should not write body")

	assert.Equal(t, r.Method, h.reqFwd.Method, "should not modify method")
	assert.Equal(t, r.URL, h.reqFwd.URL, "should not modify url")
	assert.Equal(t, r.Header, h.reqFwd.Header, "should not modify url")
	assert.Equal(t, r.Body, h.reqFwd.Body, "should not modify body")

	require.Equal(t, err, nil, "should pass error from next middleware")
	assert.Equal(t, i, 201, "should pass return code from next middleware")

	require.Len(t, h.reqGitea, 1, "should call gitea API exactly once")
	reqG := h.reqGitea[0]
	assert.Equal(t, "/api/v1/user", reqG.URL.Path, "should call GetMyInfo api")
	assert.Equal(t, http.MethodGet, reqG.Method, "should call GetMyInfo api")
	u, p, ok := reqG.BasicAuth()
	assert.Equal(t, "user", u, "should pass username")
	assert.Equal(t, "pass", p, "should pass password")
	assert.True(t, ok, "should pass basic auth")

}

func TestChi(t *testing.T) {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("middleware")
			next.ServeHTTP(w, r)
		})
	})
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(chi.RouteContext(r.Context()).RoutePattern())
	}))
	r.Handle("/{test}/{info}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(chi.RouteContext(r.Context()).RoutePattern())
	}))
	r.Handle("/test*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(chi.RouteContext(r.Context()).RoutePattern())
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	r.ServeHTTP(w, req)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/tes/info", nil)
	r.ServeHTTP(w, req)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/test/info", nil)
	r.ServeHTTP(w, req)
}
