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

type projectLogListing struct {
	Name   string
	Active bool
}

func indexPageHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", nil)
}

func githubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	webhook.GithubWebhookHandler(r)
}

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
	logFile := logFilePathFromRequest("/ws/", r)
	go writer(ws, lastMod, logFile)
}

func runCIHandler(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
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
	updateBuildStatusFunc := newGithubClientFromSession(session).UpdateBuildStatus(payload)

	webhook.ManualTrigger(payload.Repo, payload.Owner, payload.Ref, lang, url, updateBuildStatusFunc)
	http.Redirect(w, r, redirectURL, 302)
}

func ciPageHandler(w http.ResponseWriter, r *http.Request) {
	logFile := logFilePathFromRequest("/ci/", r)
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

	session, _ := sessionStore.Get(r, sessionName)
	projectPath, _ := filepath.Rel("/ci/", r.URL.Path)
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

func dashboardPageHandler(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	tkn := accessTokenFromSession(session)
	repos := getUserProjectsWithSubscriptionInfo(tkn, ghCallbackURL(r.Host))
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

func showPageHandler(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	project := r.URL.Query().Get("project")
	owner := r.URL.Query().Get("owner")
	logDir := filepath.Join(ci.LogDIR, owner, project)

	logs := listProjectLogsInDir(logDir)
	info := struct{ Logs []projectLogListing }{logs}
	renderTemplate(w, "show", info)
}

func ghSubscribe(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	project := r.URL.Query().Get("project")
	owner := r.URL.Query().Get("owner")
	redirPath := "/dashboard"

	payload := vcs.GithubRequestParams{
		Owner:       owner,
		Repo:        project,
		CallbackURL: ghCallbackURL(r.Host),
		Creds:       os.Getenv("GITHUB_WEBHOOK_SECRET"),
	}

	err := newGithubClientFromSession(session).Subscribe(payload)
	if err != nil {
		log.Println("Error while creating webhook", err)
		session.AddFlash("An error occurred. The project might have already been subscribed.")
	} else {
		session.AddFlash("Sicro is now watching: ", project)
		redirPath = fmt.Sprintf("/show?project=%s&owner=%s", project, owner)
	}

	session.Save(r, w)
	http.Redirect(w, r, redirPath, 302)
}
