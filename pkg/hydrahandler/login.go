package hydrahandler

import (
	"fmt"
	"net/http"

	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
	"golang.org/x/oauth2"
)

func (c *handler) showLogin(w http.ResponseWriter, r *http.Request) {
	lc := r.URL.Query().Get("login_challenge")
	_, err := c.hydra.Admin.GetLoginRequest(admin.NewGetLoginRequestParams().
		WithContext(r.Context()).
		WithLoginChallenge(lc))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "verify login challenge failed: %v", err)
		return
	}
	// We won't skip login so that oauth2 token always available when accepting consent.
	// This make us can't process `prompt` parameter correctly.
	// But it was not possible in the first place since we can't control gitea OAuth2 login or consent prompt
	http.Redirect(w, r, c.giteaOauth2.AuthCodeURL(lc, oauth2.AccessTypeOffline), http.StatusTemporaryRedirect)
}

func (c *handler) acceptLogin(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	token, err := c.giteaOauth2.Exchange(c.giteaOAuth2Context(r.Context()), code)
	if err != nil {
		res, err := c.hydra.Admin.RejectLoginRequest(admin.NewRejectLoginRequestParams().
			WithContext(r.Context()).
			WithLoginChallenge(state).
			WithBody(&models.RejectRequest{
				Error:            "login_required",
				ErrorDescription: err.Error(),
			}))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "reject login request failed: %v", err)
			return
		}
		http.Redirect(w, r, res.Payload.RedirectTo, http.StatusTemporaryRedirect)
		return
	}
	giteac := c.giteaOAuth2Client(r.Context(), token)

	user, err := giteac.GetMyUserInfo()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "fetch gitea user info failed: %v", err)
		return
	}

	res, err := c.hydra.Admin.AcceptLoginRequest(admin.NewAcceptLoginRequestParams().
		WithContext(r.Context()).
		WithLoginChallenge(state).
		WithBody(&models.AcceptLoginRequest{
			Subject:     &user.UserName,
			Remember:    c.hydraRemember,
			RememberFor: c.hydraRememberFor,
			Context:     loginContext{Token: token},
		}))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "accept login request failed: %v", err)
		return
	}
	http.Redirect(w, r, res.Payload.RedirectTo, http.StatusTemporaryRedirect)
}
