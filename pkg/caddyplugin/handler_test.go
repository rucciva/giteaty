package caddyplugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"

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

	reqGitea []*http.Request
	reqFwd   *http.Request
}

func insertURL(t *testing.T, conf string, url string) string {
	tmpl, err := template.New("conf").Parse(conf)
	require.NoError(t, err)

	type data struct {
		URL string
	}
	sb := &strings.Builder{}
	require.NoError(t, tmpl.Execute(sb, data{URL: url}))
	return sb.String()
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

	i := insertURL(t, conf, s.URL)
	c := caddy.NewTestController("http", i)
	require.NoError(t, setup(c), i)

	next := httpserver.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (i int, err error) {
		h.reqFwd = r
		return 201, nil
	})
	h.handler, _ = httpserver.GetConfig(c).Middleware()[0](next).(*handler)
	require.NotNil(t, h.handler)
	return
}

func (h testHandler) mustAssertUser(t *testing.T, req *http.Request) {
	assert.Equal(t, "/api/v1/user", req.URL.Path, "should call GetMyInfo api")
	assert.Equal(t, http.MethodGet, req.Method, "should call GetMyInfo api")
}

func (h testHandler) mustAssertRepo(t *testing.T, req *http.Request, owner string, repo string) {
	p := fmt.Sprintf("/api/v1/repos/%s/%s", owner, repo)
	assert.Equal(t, p, req.URL.Path, "should call get repo api")
	assert.Equal(t, http.MethodGet, req.Method, "should call get repo api")
}

func (h testHandler) mustAssertRepoBranch(t *testing.T, req *http.Request, owner string, repo string) {
	p := fmt.Sprintf("/api/v1/repos/%s/%s/branches", owner, repo)
	assert.Equal(t, p, req.URL.Path, "should call get repo branches api")
	assert.Equal(t, http.MethodGet, req.Method, "should call get repo branches api")
}

func (h testHandler) mustAssertOrgTeams(t *testing.T, req *http.Request) {
	assert.Equal(t, "/api/v1/user/teams", req.URL.Path, "should call get repo branches api")
	assert.Equal(t, http.MethodGet, req.Method, "should call get repo branches api")
}

func (h testHandler) mustUseBasicAuth(t *testing.T, req *http.Request, user string, pass string) {
	u, p, ok := req.BasicAuth()
	assert.Equal(t, user, u, "should pass username")
	assert.Equal(t, pass, p, "should pass password")
	assert.True(t, ok, "should pass basic auth")
}

func (h testHandler) mustNotWriteResponse(t *testing.T, w *httptest.ResponseRecorder) {
	assert.Equal(t, w.Code, 0, "should not write status code")
	assert.Empty(t, w.Header(), "should not write header")
	assert.Nil(t, w.Body, "should not write body")
}

func (h testHandler) mustForwardReq(t *testing.T, r *http.Request) {
	assert.Equal(t, r.Method, h.reqFwd.Method, "should not modify method")
	assert.Equal(t, r.URL, h.reqFwd.URL, "should not modify url")
	assert.Equal(t, r.Header, h.reqFwd.Header, "should not modify url")
	assert.Equal(t, r.Body, h.reqFwd.Body, "should not modify body")
}

func (h testHandler) mustForwardReqWithCustomBasicAuthPass(t *testing.T, r *http.Request, pass string) {
	uo, _, _ := r.BasicAuth()
	u, p, ok := h.reqFwd.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, uo, u)
	assert.Equal(t, pass, p)

	r = r.Clone(r.Context())
	r.SetBasicAuth(u, p)
	h.mustForwardReq(t, r)
}

func (h testHandler) mustRelayNextResponse(t *testing.T, i int, err error) {
	require.Equal(t, err, nil, "should pass error from next middleware")
	assert.Equal(t, i, 201, "should pass return code from next middleware")
}

func (h testHandler) mustNotForwardRequest(t *testing.T, i int, err error) {
	assert.Equal(t, 403, i)
	assert.Error(t, err)
	assert.Nil(t, h.reqFwd)
}

func TestHandlerPassthrough(t *testing.T) {
	conf := `
		gitea-auth /test {{.URL}}
		gitea-auth /dev {{.URL}}
	`
	ress := []testResponse{}

	paths := []string{
		"/",
		"/tester",
		"/tester/something",
		"/tester/something/else",
		"/developer",
		"/developer/something",
		"/developer/something/else",
	}

	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			w := &httptest.ResponseRecorder{}
			r := httptest.NewRequest(http.MethodGet, p, strings.NewReader(`{"message" : "hi"}`))
			h := tNewHandler(t, conf, ress)
			i, err := h.handler.ServeHTTP(w, r.Clone(r.Context()))
			h.mustRelayNextResponse(t, i, err)
			h.mustNotWriteResponse(t, w)
			h.mustForwardReq(t, r)
			require.Len(t, h.reqGitea, 0, "should not call gitea API")
		})
	}
}

func TestHandlerAssertUser(t *testing.T) {
	conf := `
	gitea-auth /test {{.URL}}
	gitea-auth /dev {{.URL}}
	`

	paths := []string{
		"/test",
		"/test/something",
		"/test/something/else",
		"/dev",
		"/dev/something",
		"/dev/something/else",
	}
	for _, p := range paths {
		t.Run("Success#"+p, func(t *testing.T) {
			user, pass := "user", "pass"
			ress := []testResponse{
				{200, nil, &gitea.User{UserName: user}},
			}

			w := &httptest.ResponseRecorder{}
			r := httptest.NewRequest(http.MethodGet, p, strings.NewReader(`{"message" : "hi"}`))
			r.SetBasicAuth(user, pass)
			h := tNewHandler(t, conf, ress)
			i, err := h.handler.ServeHTTP(w, r.Clone(r.Context()))

			h.mustRelayNextResponse(t, i, err)
			h.mustNotWriteResponse(t, w)
			h.mustForwardReq(t, r)
			require.Len(t, h.reqGitea, 1, "should call gitea API exactly once")
			h.mustAssertUser(t, h.reqGitea[0])
			h.mustUseBasicAuth(t, h.reqGitea[0], user, pass)
		})
	}

	for _, p := range paths {
		t.Run("Unauthorized#"+p, func(t *testing.T) {
			user, pass := "user", "pass"
			ress := []testResponse{
				{401, nil, map[string]interface{}{"message": "Unauthorized"}},
			}

			w := &httptest.ResponseRecorder{}
			r := httptest.NewRequest(http.MethodGet, p, strings.NewReader(`{"message" : "hi"}`))
			r.SetBasicAuth(user, pass)
			h := tNewHandler(t, conf, ress)
			i, err := h.handler.ServeHTTP(w, r.Clone(r.Context()))

			h.mustNotForwardRequest(t, i, err)
			h.mustNotWriteResponse(t, w)
			require.Len(t, h.reqGitea, 1, "should call gitea API exactly once")
			h.mustAssertUser(t, h.reqGitea[0])
			h.mustUseBasicAuth(t, h.reqGitea[0], user, pass)
		})
	}

}

func TestHandlerAssertRepo(t *testing.T) {
	conf := `
		gitea-auth /dev {{.URL}} {
			repo
		}
		gitea-auth /test {{.URL}} {
			repo /something/{owner}/{else}/{repo}.js
		}
		gitea-auth /qa {{.URL}} {
			repo sowner/srepo
		}
	`

	data := []struct {
		path  string
		owner string
		repo  string
	}{
		{"/dev/theowner/therepo", "theowner", "therepo"},
		{"/dev/theowner1/therepo1", "theowner1", "therepo1"},
		{"/dev/theowner1/therepo1/lala/lala", "theowner1", "therepo1"},

		{"/test/something/theowner/else/therepo.js", "theowner", "therepo"},
		{"/test/something/theowner/else1/therepo.js", "theowner", "therepo"},
		{"/test/something/theowner/else2/therepo.js", "theowner", "therepo"},
		{"/test/something/theowner/else2/therepo.js/lala/lala", "theowner", "therepo"},

		{"/qa", "sowner", "srepo"},
		{"/qa/lalal/syalala", "sowner", "srepo"},
	}

	for _, d := range data {
		t.Run("Success#"+d.path, func(t *testing.T) {
			user, pass := "user", "pass"
			ress := []testResponse{
				{200, nil, &gitea.Repository{Name: d.repo, Owner: &gitea.User{UserName: d.owner}}},
				{200, nil, []gitea.Branch{}},
			}

			w := &httptest.ResponseRecorder{}
			r := httptest.NewRequest(http.MethodGet, d.path, strings.NewReader(`{"message" : "hi"}`))
			r.SetBasicAuth(user, pass)
			h := tNewHandler(t, conf, ress)

			i, err := h.handler.ServeHTTP(w, r.Clone(r.Context()))
			h.mustRelayNextResponse(t, i, err)
			h.mustNotWriteResponse(t, w)
			h.mustForwardReq(t, r)
			require.Len(t, h.reqGitea, 2, "should call gitea API twice")
			h.mustAssertRepo(t, h.reqGitea[0], d.owner, d.repo)
			h.mustUseBasicAuth(t, h.reqGitea[0], user, pass)
			h.mustAssertRepoBranch(t, h.reqGitea[1], d.owner, d.repo)
			h.mustUseBasicAuth(t, h.reqGitea[1], user, pass)
		})
	}

	for _, d := range data {
		t.Run("Unauthorized#"+d.path, func(t *testing.T) {
			user, pass := "user", "pass"
			ress := []testResponse{
				{404, nil, map[string]interface{}{"message": "Not Found"}},
			}

			w := &httptest.ResponseRecorder{}
			r := httptest.NewRequest(http.MethodGet, d.path, strings.NewReader(`{"message" : "hi"}`))
			r.SetBasicAuth(user, pass)
			h := tNewHandler(t, conf, ress)

			i, err := h.handler.ServeHTTP(w, r.Clone(r.Context()))
			h.mustNotForwardRequest(t, i, err)
			h.mustNotWriteResponse(t, w)
			require.Len(t, h.reqGitea, 1, "should call gitea API once")
			h.mustAssertRepo(t, h.reqGitea[0], d.owner, d.repo)
			h.mustUseBasicAuth(t, h.reqGitea[0], user, pass)
		})
	}

	for _, d := range data {
		t.Run("UnauthorizedRepoCodeAccess#"+d.path, func(t *testing.T) {
			user, pass := "user", "pass"
			ress := []testResponse{
				{200, nil, &gitea.Repository{Name: d.repo, Owner: &gitea.User{UserName: d.owner}}},
				{403, nil, map[string]interface{}{"message": "Unauthroized"}},
			}

			w := &httptest.ResponseRecorder{}
			r := httptest.NewRequest(http.MethodGet, d.path, strings.NewReader(`{"message" : "hi"}`))
			r.SetBasicAuth(user, pass)
			h := tNewHandler(t, conf, ress)

			i, err := h.handler.ServeHTTP(w, r.Clone(r.Context()))
			h.mustNotForwardRequest(t, i, err)
			h.mustNotWriteResponse(t, w)
			require.Len(t, h.reqGitea, 2, "should call gitea API once")
			h.mustAssertRepo(t, h.reqGitea[0], d.owner, d.repo)
			h.mustUseBasicAuth(t, h.reqGitea[0], user, pass)
			h.mustAssertRepoBranch(t, h.reqGitea[1], d.owner, d.repo)
			h.mustUseBasicAuth(t, h.reqGitea[1], user, pass)
		})
	}
}

func TestHandlerAssertOrg(t *testing.T) {
	conf := `
		gitea-auth /dev {{.URL}} {
			org
		}
		gitea-auth /test {{.URL}} {
			org /something/{org}
		}
		gitea-auth /ops {{.URL}} {
			org /something/{org} {
				teams dev sec ops
			}
		}
		gitea-auth /qa {{.URL}} {
			org someorg1
		}
	`

	data := []struct {
		path  string
		org   string
		teams []string
	}{
		{"/dev/someorg", "someorg", []string{"Owners"}},

		{"/test/something/someorg", "someorg", []string{"Owners"}},

		{"/ops/something/someorg", "someorg", []string{"dev"}},
		{"/ops/something/someorg", "someorg", []string{"ops"}},
		{"/ops/something/someorg", "someorg", []string{"sec"}},

		{"/qa", "someorg1", []string{"owners"}},
		{"/qa/syalala/lala", "someorg1", []string{"owners"}},
	}
	for _, d := range data {
		t.Run("Success#"+d.path, func(t *testing.T) {
			user, pass := "user", "pass"
			teams := []*gitea.Team{}
			for _, team := range d.teams {
				t := &gitea.Team{Name: team, Organization: &gitea.Organization{UserName: d.org}}
				teams = append(teams, t)
			}
			ress := []testResponse{{200, nil, teams}}

			w := &httptest.ResponseRecorder{}
			r := httptest.NewRequest(http.MethodGet, d.path, strings.NewReader(`{"message" : "hi"}`))
			r.SetBasicAuth(user, pass)
			h := tNewHandler(t, conf, ress)

			i, err := h.handler.ServeHTTP(w, r.Clone(r.Context()))
			h.mustRelayNextResponse(t, i, err)
			h.mustNotWriteResponse(t, w)
			h.mustForwardReq(t, r)
			require.Len(t, h.reqGitea, 1, "should call gitea API once")
			h.mustAssertOrgTeams(t, h.reqGitea[0])
			h.mustUseBasicAuth(t, h.reqGitea[0], user, pass)
		})
	}

	for _, d := range data {
		t.Run("Unauthorized#"+d.path, func(t *testing.T) {
			user, pass := "user", "pass"
			ress := []testResponse{{404, nil, map[string]interface{}{"message": "Not Found"}}}

			w := &httptest.ResponseRecorder{}
			r := httptest.NewRequest(http.MethodGet, d.path, strings.NewReader(`{"message" : "hi"}`))
			r.SetBasicAuth(user, pass)
			h := tNewHandler(t, conf, ress)

			i, err := h.handler.ServeHTTP(w, r.Clone(r.Context()))
			h.mustNotForwardRequest(t, i, err)
			h.mustNotWriteResponse(t, w)
			require.Len(t, h.reqGitea, 1, "should call gitea API once")
			h.mustAssertOrgTeams(t, h.reqGitea[0])
			h.mustUseBasicAuth(t, h.reqGitea[0], user, pass)

		})
	}

	for _, d := range data {
		t.Run("NotMemberOf#"+d.path, func(t *testing.T) {
			user, pass := "user", "pass"
			ress := []testResponse{
				{200, nil, []*gitea.Team{{Name: "noone", Organization: &gitea.Organization{UserName: d.org}}}},
			}

			w := &httptest.ResponseRecorder{}
			r := httptest.NewRequest(http.MethodGet, d.path, strings.NewReader(`{"message" : "hi"}`))
			r.SetBasicAuth(user, pass)
			h := tNewHandler(t, conf, ress)

			i, err := h.handler.ServeHTTP(w, r.Clone(r.Context()))
			h.mustNotForwardRequest(t, i, err)
			h.mustNotWriteResponse(t, w)
			require.Len(t, h.reqGitea, 1, "should call gitea API once")
			h.mustAssertOrgTeams(t, h.reqGitea[0])
			h.mustUseBasicAuth(t, h.reqGitea[0], user, pass)
		})
	}
}

func TestSetBasicAuth(t *testing.T) {
	conf := `
		gitea-auth /dev {{.URL}} {
			setBasicAuth syalala
			repo /repo/{owner}/{repo}
			org /org/{org}
		}
	`

	paths := []string{
		"/dev",
		"/dev/test",
		"/dev/ops",
	}
	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			user, pass := "user", "pass"
			ress := []testResponse{
				{200, nil, &gitea.User{UserName: user}},
			}

			w := &httptest.ResponseRecorder{}
			r := httptest.NewRequest(http.MethodGet, p, strings.NewReader(`{"message" : "hi"}`))
			r.SetBasicAuth(user, pass)
			h := tNewHandler(t, conf, ress)
			i, err := h.handler.ServeHTTP(w, r.Clone(r.Context()))

			h.mustRelayNextResponse(t, i, err)
			h.mustNotWriteResponse(t, w)
			h.mustForwardReqWithCustomBasicAuthPass(t, r, "syalala")
			require.Len(t, h.reqGitea, 1, "should call gitea API exactly once")
			h.mustAssertUser(t, h.reqGitea[0])
			h.mustUseBasicAuth(t, h.reqGitea[0], user, pass)
		})
	}

}
