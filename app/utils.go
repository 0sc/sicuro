package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/0sc/sicuro/app/vcs"
	"github.com/0sc/sicuro/ci"
	"github.com/google/go-github/github"
	"github.com/gorilla/sessions"
)

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".tmpl", data)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func accessTokenFromSession(session *sessions.Session) string {
	return session.Values["accessToken"].(string)
}

func logFilePathFromRequest(prefix string, r *http.Request) string {
	path, _ := filepath.Rel(prefix, r.URL.Path)
	return fmt.Sprintf("%s%s", filepath.Join(ci.LogDIR, path), ci.LogFileExt)
}

func fetchTemplates() (templates []string) {
	templateFolderName := "templates"
	templateFolder := filepath.Join(appDIR, templateFolderName)
	folder, err := os.Open(templateFolder)
	if err != nil {
		log.Println("Error opening template folder", err)
		return
	}

	files, err := folder.Readdir(0)
	if err != nil {
		log.Println("Error reading from template folder", err)
		return
	}

	for _, file := range files {
		templates = append(templates, filepath.Join(appDIR, templateFolderName, file.Name()))
	}
	return templates
}

func listProjectLogsInDir(dirName string) []projectLogListing {
	fileInfo, err := os.Stat(dirName)
	logs := []projectLogListing{}
	if err != nil {
		log.Printf("An error: %s; occured with listing logs for %s\n", err, dirName)
		return logs
	}
	if !fileInfo.IsDir() {
		return append(logs, projectLogListing{Name: dirName, Active: ci.ActiveCISession(dirName)})
	}
	dir, err := os.Open(dirName)
	if err != nil {
		log.Printf("An error: %s; occured will opening dir %s", err, dirName)
		return logs
	}
	files, err := dir.Readdir(0)
	if err != nil {
		log.Printf("An error: %s; occured will reading files from dir %s", err, dirName)
		return logs
	}
	for _, file := range files {
		fileFullName := filepath.Join(dirName, file.Name())
		if file.IsDir() {
			logs = append(logs, listProjectLogsInDir(fileFullName+"/")...)
		} else {
			name, _ := filepath.Rel(ci.LogDIR, fileFullName)
			name = strings.Replace(name, ci.LogFileExt, "", 1)
			logs = append(logs, projectLogListing{Name: name, Active: ci.ActiveCISession(fileFullName)})
		}
	}
	return logs
}

func getUserProjectsWithSubscriptionInfo(token, webhookPath string) []RepoWithSubscriptionInfo {
	client := vcs.NewGithubClient(token)
	repos := []RepoWithSubscriptionInfo{}
	params := vcs.GithubRequestParams{CallbackURL: webhookPath}

	for _, repo := range client.UserRepos() {
		params.Owner = *(repo.Owner.Login)
		params.Repo = *(repo.Name)

		repos = append(repos, RepoWithSubscriptionInfo{client.IsRepoSubscribed(params), repo})
	}

	return repos
}

func getProject(token, owner, project string) (*github.Repository, error) {
	payload := vcs.GithubRequestParams{
		Owner: owner,
		Repo:  project,
	}
	return vcs.NewGithubClient(token).Repo(payload)
}

func newGithubClientFromSession(session *sessions.Session) *vcs.GithubClient {
	return vcs.NewGithubClient(accessTokenFromSession(session))
}

func ghCallbackURL(hostAddr string) string {
	return fmt.Sprintf("http://%s/gh/webhook", hostAddr)
}
