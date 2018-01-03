package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	redirectURL        = ""
)

var (
	githubOauthCfg *oauth2.Config
	scopes         = []string{"repo"}
)

type RepoWithSubscriptionInfo struct {
	IsSubscribed bool
	*github.Repository
}

func setupGithubOauthCfg() {
	githubOauthCfg = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  githubAuthorizeURL,
			TokenURL: githubTokenURL,
		},
		RedirectURL: redirectURL,
		Scopes:      scopes,
	}
}

func ghAuth(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	rand.Read(b)

	state := base64.URLEncoding.EncodeToString(b)
	session, err := fetchSession(r)
	if err != nil {
		panic(err)
	}
	session.Values["state"] = state
	err = session.Save(r, w)

	if err != nil {
		panic(err)
	}

	url := githubOauthCfg.AuthCodeURL(state)
	http.Redirect(w, r, url, 302)
}

func ghAuthCallback(w http.ResponseWriter, r *http.Request) {
	session, err := fetchSession(r)

	if err != nil {
		fmt.Fprintln(w, "aborted")
		return
	}

	if r.URL.Query().Get("state") != session.Values["state"] {
		fmt.Fprintln(w, "no state match; possible csrf OR cookies not enabled")
		return
	}

	tkn, err := githubOauthCfg.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
	if err != nil {
		fmt.Fprintln(w, "there was an issue getting your token")
		return
	}

	if !tkn.Valid() {
		fmt.Fprintln(w, "retrieved invalid token")
		return
	}

	client := github.NewClient(githubOauthCfg.Client(oauth2.NoContext, tkn))

	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		fmt.Println(w, "error getting name")
		return
	}

	session.Values["name"] = user.Name
	session.Values["accessToken"] = tkn.AccessToken
	session.Save(r, w)

	http.Redirect(w, r, "/dashboard", 302)
}
