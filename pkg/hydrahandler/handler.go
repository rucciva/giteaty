package hydrahandler

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"

	"code.gitea.io/sdk/gitea"
	"github.com/go-chi/chi"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/ory/hydra-client-go/client"
	"golang.org/x/oauth2"
)

type option func(*handler) error

func WithHydraAdminURL(u string) option {
	return func(c *handler) (err error) {
		c.hydraAdminURL = u
		return
	}
}

func WithGiteaURL(u string) option {
	return func(c *handler) (err error) {
		gurl, err := url.Parse(u)
		if err != nil {
			return
		}

		c.giteaURL = gurl.Scheme + "://" + gurl.Host
		c.giteaOauth2.Endpoint = oauth2.Endpoint{
			AuthURL:  c.giteaURL + "/login/oauth/authorize",
			TokenURL: c.giteaURL + "/login/oauth/access_token",
		}
		return
	}
}

func WithGiteaOauth2Client(id, secret, redirectURI string) option {
	return func(c *handler) (err error) {
		rurl, err := url.Parse(redirectURI)
		if err != nil {
			return
		}
		c.giteaOauth2.ClientID = id
		c.giteaOauth2.ClientSecret = secret
		c.giteaOauth2.RedirectURL = rurl.String()
		return
	}
}

func WithAllowInsecure() option {
	return func(c *handler) (err error) {
		c.allowInsecure = true
		return
	}
}

func WithBaseURL(u string) option {
	return func(c *handler) (err error) {
		burl, err := url.Parse(u)
		if err != nil {
			return
		}
		c.baseUrl = burl.Scheme + "://" + burl.Host
		return
	}
}

type handler struct {
	baseUrl string

	giteaURL    string
	giteaOauth2 *oauth2.Config
	hclient     *http.Client

	hydraAdminURL    string
	hydra            *client.OryHydra
	hydraRemember    bool
	hydraRememberFor int64

	allowInsecure bool
}

func New(opts ...option) (c *handler, err error) {
	c = &handler{
		giteaOauth2: &oauth2.Config{},

		hydra:         client.Default,
		hydraRemember: true,
	}

	for _, opt := range opts {
		if err = opt(c); err != nil {
			return
		}
	}

	// http client
	c.hclient = &http.Client{}
	if c.allowInsecure {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		c.hclient.Transport = tr
	}

	// hydra admin client
	adminURL, err := url.Parse(c.hydraAdminURL)
	if err != nil {
		return
	}
	tr := httptransport.New(adminURL.Host, adminURL.Path, []string{adminURL.Scheme})
	tr.Transport = c.hclient.Transport
	c.hydra = client.New(tr, nil)

	return
}

func (c *handler) Handler() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/hydra/login", c.showLogin)
	r.Get("/hydra/consent", c.acceptConsent)
	r.Get("/hydra/logout", c.showLogout)
	r.Get("/hydra/login/callback", c.acceptLogin)
	r.Get("/hydra/logout/callback", c.acceptLogout)
	r.Get("/gitea/logout", c.showGiteaLogout)
	return r
}

func (c *handler) giteaOAuth2Context(ctx context.Context) context.Context {
	if !c.allowInsecure {
		return ctx
	}
	return context.WithValue(ctx, oauth2.HTTPClient, c.hclient)
}

func (c *handler) giteaOAuth2Client(ctx context.Context, t *oauth2.Token) *gitea.Client {
	hcli := c.giteaOauth2.Client(c.giteaOAuth2Context(ctx), t)
	return gitea.NewClientWithHTTP(c.giteaURL, hcli)
}
