package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/0sc/sicuro/app/vcs"
	"github.com/0sc/sicuro/app/webhook"
	"github.com/0sc/sicuro/ci"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

var (
	sessionSecret = os.Getenv("SESSION_SECRET")
	sessionStore  = sessions.NewCookieStore([]byte(sessionSecret))
	templates     = template.Must(template.ParseFiles(fetchTemplates()...))
	upgrader      = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	var lastMod time.Time
	if n, err := strconv.ParseInt(r.FormValue("lastMod"), 16, 64); err == nil {
		lastMod = time.Unix(0, n)
	}
	logFile := logFilePathFromRequest(websocketPath, r)
	go writer(ws, lastMod, logFile)
}

func githubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	webhook.GithubWebhookHandler(r)
}

func githubSubscriptionHandler() http.HandlerFunc {
	self := func(w http.ResponseWriter, r *http.Request) {
		project := r.URL.Query().Get("project")
		owner := r.URL.Query().Get("owner")
		redirPath := dashboardPath
		session, _ := fetchSession(r)
		token := r.Context().Value(accessTokenCtxKey).(string)
		client := newGithubClient(token)

		payload := vcs.GithubRequestParams{
			Owner:       owner,
			Repo:        project,
			CallbackURL: ghCallbackURL(r.Host),
			Creds:       os.Getenv("GITHUB_WEBHOOK_SECRET"),
		}

		err := client.Subscribe(payload)
		if err != nil {
			log.Println("Error while creating webhook", err)
			session.AddFlash("An error occurred. The project might have already been subscribed.")
		} else {
			session.AddFlash("Sicro is now watching: ", project)
			redirPath = fmt.Sprintf("%s?project=%s&owner=%s", showPath, project, owner)
		}

		session.Save(r, w)
		http.Redirect(w, r, redirPath, http.StatusTemporaryRedirect)
	}

	middlewares := []middleware{
		authenticationMiddleware,
	}

	return buildMiddlewareChain(self, middlewares...)
}

func ciPageHandler() http.HandlerFunc {
	self := func(w http.ResponseWriter, r *http.Request) {
		logFile := logFilePathFromRequest(ciPath, r)
		fmt.Println("The logfile", logFile)
		if _, err := os.Open(logFile); err != nil {
			http.Error(w, "Not found", 404)
			return
		}
		p, lastMod, err := readFileIfModified(logFile, time.Time{})
		if err != nil {
			p = []byte(err.Error())
			lastMod = time.Unix(0, 0)
		}

		session, _ := fetchSession(r)
		projectPath, _ := filepath.Rel(ciPath, r.URL.Path)
		details := strings.Split(projectPath, "/")

		var v = struct {
			Owner         string
			Project       string
			Commit        string
			Host          string
			Data          template.HTML
			LastMod       string
			ProjectPath   string
			Notifications []interface{}
		}{
			Owner:         details[0],
			Project:       details[1],
			Commit:        details[2],
			Host:          r.Host,
			Data:          template.HTML(p),
			LastMod:       strconv.FormatInt(lastMod.UnixNano(), 16),
			ProjectPath:   projectPath,
			Notifications: session.Flashes(),
		}
		renderTemplate(w, "ci", &v)
	}

	middlewares := []middleware{
		validateRequestMethod("GET"),
	}

	return buildMiddlewareChain(self, middlewares...)
}

func dashboardPageHandler() http.HandlerFunc {
	self := func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value(accessTokenCtxKey).(string)
		repos := getUserProjectsWithSubscriptionInfo(token, ghCallbackURL(r.Host))
		session, _ := fetchSession(r)

		info := struct {
			FlashMsgs []interface{}
			Repos     []RepoWithSubscriptionInfo
		}{
			FlashMsgs: session.Flashes(),
			Repos:     repos,
		}
		session.Save(r, w)
		renderTemplate(w, "dashboard", info)
	}

	middlewares := []middleware{
		validateRequestMethod("GET"),
		authenticationMiddleware,
	}

	return buildMiddlewareChain(self, middlewares...)
}

func indexPageHandler() http.HandlerFunc {
	self := func(w http.ResponseWriter, r *http.Request) {
		_, err := fetchSession(r)
		if err == nil { // then user has a session set
			http.Redirect(w, r, dashboardPath, http.StatusTemporaryRedirect)
			return
		}
		renderTemplate(w, "index", nil)
	}

	middlewares := []middleware{
		validateRequestMethod("GET"),
	}

	return buildMiddlewareChain(self, middlewares...)
}

func showPageHandler() http.HandlerFunc {
	self := func(w http.ResponseWriter, r *http.Request) {
		project := r.URL.Query().Get("project")
		owner := r.URL.Query().Get("owner")
		logDir := filepath.Join(ci.LogDIR, owner, project)

		logs := listProjectLogsInDir(logDir)
		info := struct {
			Logs []projectLogListing
		}{logs}
		renderTemplate(w, "show", info)
	}

	middlewares := []middleware{
		validateRequestMethod("GET"),
		authenticationMiddleware,
		authorizationMiddleware,
		projectSubscriptionMiddleware,
	}

	return buildMiddlewareChain(self, middlewares...)
}

func runCIHandler() http.HandlerFunc {
	self := func(w http.ResponseWriter, r *http.Request) {
		repo := r.URL.Query().Get("repo")
		redirectURL := fmt.Sprintf("ci/%s", repo)

		payload := vcs.GithubRequestParams{
			Repo:        r.URL.Query().Get("project"),
			Owner:       r.URL.Query().Get("owner"),
			Ref:         r.URL.Query().Get("sha"),
			CallbackURL: fmt.Sprintf("http://%s/%s", r.Host, redirectURL),
		}

		lang := r.URL.Query().Get("language")
		url := r.URL.Query().Get("url")
		token := r.Context().Value(accessTokenCtxKey).(string)

		updateBuildStatusFunc := newGithubClient(token).UpdateBuildStatus(payload)

		webhook.ManualTrigger(payload.Repo, payload.Owner, payload.Ref, lang, url, updateBuildStatusFunc)
		http.Redirect(w, r, redirectURL, 302)
	}

	middlewares := []middleware{
		validateRequestMethod("GET"),
		parseProjectDetailsMiddleware,
		authenticationMiddleware,
		authorizationMiddleware,
		projectSubscriptionMiddleware,
	}

	return buildMiddlewareChain(self, middlewares...)
}
