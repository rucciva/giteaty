package caddyhandler

import (
	"crypto/tls"
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
)

func (drt *Directive) newGiteaClient(req *http.Request) *gitea.Client {
	rt := roundtripper{RoundTripper: http.DefaultTransport, caddyReq: req}
	if drt.insecure {
		rt.RoundTripper = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	hc := &http.Client{Transport: rt}
	gc := gitea.NewClientWithHTTP(drt.giteaURL, hc)
	return gc
}

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
	if u, p, ok := rt.caddyReq.BasicAuth(); ok {
		// otp workaround
		if s := strings.Split(u, ";"); len(s) == 2 {
			req.SetBasicAuth(s[0], p)
			req.Header[http.CanonicalHeaderKey("X-Gitea-OTP")] = []string{s[1]}
		}
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
