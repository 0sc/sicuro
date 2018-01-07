package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
)

var (
	githubOAuth *oauth2.Config
	scopes      = []string{"repo"}
)

func setupGithubOAuth() {
	githubOAuth = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  githubAuthorizeURL,
			TokenURL: githubTokenURL,
		},
		Scopes: scopes,
	}
}

func ghAuthHandler(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	rand.Read(b)

	state := base64.URLEncoding.EncodeToString(b)
	session, err := fetchSession(r)
	if err != nil {
		log.Println("Error occured while fetching session: ", err)
		renderTemplate(w, "error", "Your browser session is invalid. Please try again.")
		return
	}

	session.Values["state"] = state
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error occured while saving state code to session: ", err)
		renderTemplate(w, "error", "We encountered an error while saving your browser session. Please try again.")
		return
	}

	url := githubOAuth.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func ghAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := fetchSession(r)
	if err != nil {
		log.Println("Error occured while fetching session: ", err)
		renderTemplate(w, "error", "Oops! Login request didn't complete successfully. Please try again.")
		return
	}

	if r.URL.Query().Get("state") != session.Values["state"] {
		renderTemplate(w, "error", "Hmm ... your login request seems fishy. Possible CSRF or maybe cookies not enabled. Please try again")
		return
	}

	tkn, err := githubOAuth.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
	if err != nil {
		log.Println("Error occured while exchanging Github Access Token: ", err)
		renderTemplate(w, "error", "We couldn't retrieve your Github Access Token. Please try again")
		return
	}

	if !tkn.Valid() {
		log.Println("Error retrieved token is invalid")
		renderTemplate(w, "error", "The token we got from Github is invalid. Please try again")
		return
	}

	// client := github.NewClient(githubOAuth.Client(oauth2.NoContext, tkn))

	// user, _, err := client.Users.Get(context.Background(), "")
	// if err != nil {
	// 	log.Println("Error occured while getting user name: ", err)
	// 	fmt.Println(w, "error getting name")
	// 	return
	// }

	// session.Values["name"] = user.Name
	session.Values[accessTokenKey] = tkn.AccessToken
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error occured while saving access token: ", err)
		renderTemplate(w, "error", "Something went wrong while handling your token. Please try again")
	}

	http.Redirect(w, r, dashboardPath, http.StatusTemporaryRedirect)
}
