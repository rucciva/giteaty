package hydrahandler

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/ory/hydra-client-go/client/admin"
)

//WARN: logout hack
var logoutHTML = `
<html>
<head>
	<title> Logout </title>
	<meta http-equiv="refresh" content="3;URL='{{.LogoutURL}}'" /> 
</head>
<body >
	<p>Please wait while you are being redirected...</p>
	<iframe src="{{.GiteaLogoutURL}}" style="width:0; height:0; border:0; border:none"></iframe>
</body>
</html>
`

//WARN: logout hack
var giteaLogoutHTML = `
<html>
<head>
	<title> Gitea Logout </title>
</head>
<body onload="document.forms['logout'].submit()">
	<form action="{{.GiteaLogoutURL}}" method="post" name="logout"></form>
</body>
</html>
`

//WARN: logout hack
func (c *handler) showLogout(w http.ResponseWriter, r *http.Request) {
	lc := r.URL.Query().Get("logout_challenge")

	tmpl, err := template.New("html").Parse(logoutHTML)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "parse template failed: %v", err)
		return
	}
	w.Header().Add("content-type", "text/html")
	err = tmpl.Execute(w, struct {
		LogoutURL      string
		GiteaLogoutURL string
	}{
		LogoutURL:      c.baseUrl + "/hydra/logout/callback?logout_challenge=" + lc,
		GiteaLogoutURL: "/gitea/logout",
	})
	if err != nil {
		fmt.Fprintf(w, "execute template failed: %v", err)
		return
	}
}

//WARN: logout hack
func (c *handler) showGiteaLogout(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("html").Parse(giteaLogoutHTML)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "parse template failed: %v", err)
		return
	}
	w.Header().Add("content-type", "text/html")
	err = tmpl.Execute(w, struct{ GiteaLogoutURL string }{
		GiteaLogoutURL: c.giteaURL + "/user/logout",
	})
	if err != nil {
		fmt.Fprintf(w, "execute template failed: %v", err)
		return
	}
}

func (c *handler) acceptLogout(w http.ResponseWriter, r *http.Request) {
	lc := r.URL.Query().Get("logout_challenge")

	res, err := c.hydra.Admin.AcceptLogoutRequest(admin.NewAcceptLogoutRequestParams().
		WithContext(r.Context()).
		WithLogoutChallenge(lc),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "verify login challenge failed: %v", err)
		return
	}
	http.Redirect(w, r, res.Payload.RedirectTo, http.StatusTemporaryRedirect)
}
