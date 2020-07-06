package caddyplugin

import (
	"crypto/tls"
	"net/http"

	"code.gitea.io/sdk/gitea"
)

type roundtripper struct {
	http.RoundTripper

	req *http.Request
}

func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header and URL
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	*r2.URL = *r.URL
	return r2
}

func (r roundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		defer req.Body.Close()
	}
	req2 := cloneRequest(req)

	req2.Header.Set("Authorization", r.req.Header.Get("Authorization"))
	req2.Header.Set("X-Gitea-OTP", r.req.Header.Get("X-Gitea-OTP"))
	req2.URL.Query().Set("access_token", r.req.URL.Query().Get("access_token"))
	req2.URL.Query().Set("token", r.req.URL.Query().Get("token"))
	return r.RoundTripper.RoundTrip(req2)
}

func (h *handler) newGiteaClient(req *http.Request) *gitea.Client {
	rt := roundtripper{RoundTripper: http.DefaultTransport, req: req}
	if h.cfg.giteaAllowInsecure {
		rt.RoundTripper = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	hc := &http.Client{Transport: rt}
	gc := gitea.NewClientWithHTTP(h.cfg.giteaURL, hc)
	return gc
}
