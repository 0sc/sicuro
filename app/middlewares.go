package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/0sc/sicuro/ci"
)

type ctxKey string
type middleware func(http.HandlerFunc) http.HandlerFunc

const accessTokenKey = "AccessToken"
const accessTokenCtxKey ctxKey = accessTokenKey

func buildMiddlewareChain(f http.HandlerFunc, m ...middleware) http.HandlerFunc {
	if len(m) == 0 {
		return f
	}
	return m[0](buildMiddlewareChain(f, m[1:cap(m)]...))
}

func validateRequestMethod(mtd string) middleware {
	mware := func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != mtd {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			f.ServeHTTP(w, r)
		}
	}

	return mware
}

func authenticationMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := fetchSession(r)
		if err != nil {
			http.Redirect(w, r, indexPath, http.StatusTemporaryRedirect)
			return
		}

		if tkn, ok := session.Values[accessTokenKey]; !ok {
			http.Redirect(w, r, ghAuthPath, 302)
		} else {
			ctx := context.WithValue(r.Context(), accessTokenCtxKey, tkn.(string))
			f.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func authorizationMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessTkn := r.Context().Value(accessTokenCtxKey).(string)
		project := r.URL.Query().Get("project")
		owner := r.URL.Query().Get("owner")

		repo, err := getProject(accessTkn, owner, project)
		if err != nil {
			flashMsg := "An error occurred while looking up the project. Please confirm that the project exists"
			addFlashMsg(flashMsg, w, r)
			http.Redirect(w, r, dashboardPath, http.StatusTemporaryRedirect)
			return
		}

		values := r.URL.Query()
		values.Add("language", *(repo.Language))
		values.Add("url", *(repo.HTMLURL))
		r.URL.RawQuery = values.Encode()

		f.ServeHTTP(w, r)
	}
}

func projectSubscriptionMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		project := r.URL.Query().Get("project")
		owner := r.URL.Query().Get("owner")
		logDir := filepath.Join(ci.LogDIR, owner, project) // TODO: move to a func in ci ci.ProjectLogExist()

		if _, err := os.Stat(logDir); err != nil {
			flashMsg := "Oops! Looks like the project is not subscribed. Please subscribe and try again."
			addFlashMsg(flashMsg, w, r)

			http.Redirect(w, r, dashboardPath, 302)
			return
		}

		f.ServeHTTP(w, r)
	}
}

// Update URL's accross the app to remove this
func parseProjectDetailsMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := r.URL.Query().Get("repo")
		details := strings.Split(repo, "/")
		attrs := []string{"owner", "project", "sha"}
		values := r.URL.Query()

		if len(details) > len(attrs) {
			http.Redirect(w, r, dashboardPath, 302)
			return
		}

		for i, val := range details {
			values.Add(attrs[i], val)
		}
		r.URL.RawQuery = values.Encode()

		f.ServeHTTP(w, r)
	}
}
