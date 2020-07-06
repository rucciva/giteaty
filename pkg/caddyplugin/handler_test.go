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

	giteaReq []*http.Request
	fwddReq  *http.Request
}

func tNewHandler(t *testing.T, conf string, ress []testResponse) (h *testHandler) {
	h = &testHandler{}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.giteaReq = append(h.giteaReq, r)

		res := ress[len(h.giteaReq)-1]
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
		h.fwddReq = r
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

	require.Equal(t, err, nil, "should pass error from next middleware")
	assert.Equal(t, i, 201, "should pass return code from next middleware")

	assert.Equal(t, w.Code, 0, "should not write status code")
	assert.Empty(t, w.Header(), "should not write header")
	assert.Nil(t, w.Body, "should not write body")

	require.Len(t, h.giteaReq, 1, "should call gitea API exactly once")
	reqG := h.giteaReq[0]
	assert.Equal(t, "/api/v1/user", reqG.URL.Path, "should call GetMyInfo api")
	assert.Equal(t, http.MethodGet, reqG.Method, "should call GetMyInfo api")

	u, p, ok := reqG.BasicAuth()
	assert.Equal(t, "user", u, "should pass username")
	assert.Equal(t, "pass", p, "should pass password")
	assert.True(t, ok, "should pass basic auth")

	reqF := h.fwddReq
	assert.Equal(t, r.URL, reqF.URL, "should not modify url")
	assert.Equal(t, r.Header, reqF.Header, "should not modify url")
	assert.Equal(t, r.Body, reqF.Body, "should not modify body")
}
