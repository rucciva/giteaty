package caddyplugin

import (
	"crypto/tls"
	"net/http"

	"code.gitea.io/sdk/gitea"
)

type roundtripper struct {
	http.RoundTripper

	caddyReq *http.Request
}

func (rt roundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		defer req.Body.Close()
	}

	req = req.Clone(req.Context())
	if v, ok := rt.caddyReq.Header[http.CanonicalHeaderKey("Authorization")]; ok {
		req.Header[http.CanonicalHeaderKey("Authorization")] = append([]string(nil), v...)
	}
	if v, ok := rt.caddyReq.Header[http.CanonicalHeaderKey("X-Gitea-OTP")]; ok {
		req.Header[http.CanonicalHeaderKey("X-Gitea-OTP")] = append([]string(nil), v...)
	}
	if v, ok := rt.caddyReq.URL.Query()["access_token"]; ok {
		req.URL.Query()["access_token"] = append([]string(nil), v...)
	}
	if v, ok := rt.caddyReq.URL.Query()["token"]; ok {
		req.URL.Query()["token"] = append([]string(nil), v...)
	}

	return rt.RoundTripper.RoundTrip(req)
}

func (drt *directive) newGiteaClient(req *http.Request) *gitea.Client {
	rt := roundtripper{RoundTripper: http.DefaultTransport, caddyReq: req}
	if drt.giteaAllowInsecure {
		rt.RoundTripper = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	hc := &http.Client{Transport: rt}
	gc := gitea.NewClientWithHTTP(drt.giteaURL, hc)
	return gc
}
