package hydrahandler

import (
	"fmt"
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
)

// TODO: show consent
func (c *handler) acceptConsent(w http.ResponseWriter, r *http.Request) {
	cc := r.URL.Query().Get("consent_challenge")
	res, err := c.hydra.Admin.GetConsentRequest(admin.NewGetConsentRequestParams().
		WithContext(r.Context()).
		WithConsentChallenge(cc),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "verify login challenge failed: %v", err)
		return
	}
	lc := loginContextFromRaw(res.Payload.Context)
	giteac := c.giteaOAuth2Client(r.Context(), lc.Token)

	user, err := giteac.GetMyUserInfo()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "fetch gitea user info failed: %v", err)
		return
	}
	tokenInfo := map[string]interface{}{"sub": user.UserName}
	idtokenInfo := map[string]interface{}{"sub": user.UserName}
	for _, sc := range res.Payload.RequestedScope {
		if strings.EqualFold("email", sc) && user.Email != "" {
			tokenInfo["email"] = user.Email
			idtokenInfo["email"] = user.Email
		}
		if strings.EqualFold("profile", sc) {
			if user.FullName != "" {
				tokenInfo["name"] = user.FullName
				idtokenInfo["name"] = user.FullName
			}
			tokenInfo["preferred_username"] = user.UserName
			if user.AvatarURL != "" {
				tokenInfo["picture"] = user.AvatarURL
				idtokenInfo["picture"] = user.AvatarURL
			}
		}
	}

	teams, err := giteac.ListMyTeams(&gitea.ListTeamsOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "fetch gitea teams info failed: %v", err)
		return
	}
	groupInfo := map[string][]string{}
	groupFlattenInfo := []string{}
	for _, team := range teams {
		if groupInfo[team.Organization.UserName] == nil {
			groupFlattenInfo = append(groupFlattenInfo, team.Organization.UserName)
		}
		groupFlattenInfo = append(groupFlattenInfo, fmt.Sprintf("%s[%s]", team.Organization.UserName, team.Name))
		groupInfo[team.Organization.UserName] = append(groupInfo[team.Organization.UserName], team.Name)
	}
	if len(groupInfo) > 0 {
		tokenInfo["groups"] = groupInfo
		tokenInfo["groups_flatten"] = groupFlattenInfo
		idtokenInfo["groups"] = groupInfo
		idtokenInfo["groups_flatten"] = groupFlattenInfo
	}

	res1, err := c.hydra.Admin.AcceptConsentRequest(admin.NewAcceptConsentRequestParams().
		WithContext(r.Context()).
		WithConsentChallenge(cc).
		WithBody(&models.AcceptConsentRequest{
			GrantScope:               res.Payload.RequestedScope,
			GrantAccessTokenAudience: res.Payload.RequestedAccessTokenAudience,
			Remember:                 c.hydraRemember,
			RememberFor:              c.hydraRememberFor,
			Session: &models.ConsentRequestSession{
				AccessToken: tokenInfo,
				IDToken:     tokenInfo,
			},
		}))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "accept consent request failed: %v", err)
		return
	}
	http.Redirect(w, r, res1.Payload.RedirectTo, http.StatusTemporaryRedirect)
}
