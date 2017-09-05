package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/0sc/sicuro/ci"
	"github.com/gorilla/sessions"
)

func ensureValidRequestMethod(cb func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		cb(w, r)
	}
}

func ensureUserAuthentication(cb func(http.ResponseWriter, *http.Request, *sessions.Session)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionStore.Get(r, sessionName)
		if err != nil {
			http.Redirect(w, r, "/index", 302)
			return
		}

		if _, ok := session.Values["accessToken"]; !ok {
			http.Redirect(w, r, "/gh/auth", 302)
			return
		}
		cb(w, r, session)
	}
}

func ensureValidProject(cb func(http.ResponseWriter, *http.Request, *sessions.Session)) func(http.ResponseWriter, *http.Request, *sessions.Session) {
	return func(w http.ResponseWriter, r *http.Request, s *sessions.Session) {
		token := accessTokenFromSession(s)
		project := r.URL.Query().Get("project")
		owner := r.URL.Query().Get("owner")
		repo, err := getProject(token, owner, project)

		if err != nil {
			s.AddFlash("An error occurred while looking up the project. Please confirm that the project exists")
			s.Save(r, w)
			http.Redirect(w, r, "/dashboard", 302)
			return
		}
		values := r.URL.Query()
		values.Add("language", *(repo.Language))
		values.Add("url", *(repo.HTMLURL))
		r.URL.RawQuery = values.Encode()

		cb(w, r, s)
	}
}

func ensureSubscribedProject(cb func(http.ResponseWriter, *http.Request, *sessions.Session)) func(http.ResponseWriter, *http.Request, *sessions.Session) {
	return func(w http.ResponseWriter, r *http.Request, s *sessions.Session) {
		project := r.URL.Query().Get("project")
		owner := r.URL.Query().Get("owner")
		logDir := filepath.Join(ci.LogDIR, owner, project)
		if _, err := os.Stat(logDir); err != nil {
			s.AddFlash("Oops! Looks like the project is not subscribed. Please subscribe and try again.")
			s.Save(r, w)
			http.Redirect(w, r, "/dashboard", 302)
			return
		}

		cb(w, r, s)
	}
}

func addProjectDetailsToParams(cb func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := r.URL.Query().Get("repo")
		details := strings.Split(repo, "/")
		attrs := []string{"owner", "project", "sha"}
		values := r.URL.Query()

		if len(details) > len(attrs) {
			http.Redirect(w, r, "/dashboard", 302)
			return
		}

		for i, val := range details {
			values.Add(attrs[i], val)
		}
		r.URL.RawQuery = values.Encode()

		cb(w, r)
	}
}
