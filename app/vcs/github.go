package vcs

import (
	"context"
	"log"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

var ctx = context.Background()

// GithubClient is a wrapper around the github.client instance.
// It allows addition of custom method to the instance
type GithubClient struct {
	*github.Client
}

// GithubRequestParams is a collection of common params
// required by github related methods in this this package
type GithubRequestParams struct {
	// Repo is the request target repo
	Repo string
	// Owner refers to the owner of the request target repo
	Owner string
	// Ref is the target sha ref on the repo
	Ref string
	// Callback is a URL linking back to the site
	CallbackURL string
	// Creds is any credential or secret required for the request
	Creds string
}

// NewGithubClient creates a new GithubClient with the given token
// The given token is passed to the underlying github.Client initialization
func NewGithubClient(token string) *GithubClient {
	tkn := &oauth2.Token{AccessToken: token}
	ts := oauth2.StaticTokenSource(tkn)
	tc := oauth2.NewClient(ctx, ts)
	return &GithubClient{github.NewClient(tc)}
}

// UpdateBuildStatus returns a function that when executed updates the repo status with the given status
// it takes the repo, owner and ref as args
func (client *GithubClient) UpdateBuildStatus(params GithubRequestParams) func(string) {
	status := &github.RepoStatus{
		TargetURL: github.String(params.CallbackURL),
		Context:   github.String("SicuroCI"),
	}

	return func(state string) {
		var description string

		switch state {
		case "success":
			description = "Your tests passed on Sicuro"
		case "pending":
			description = "Sicuro is running your tests"
		case "failure":
			description = "Your tests failed on Sicuro"
		case "error":
			description = "Sicuro couldn't run your tests. An error occurred"
		}

		status.State = github.String(state)
		status.Description = github.String(description)
		_, _, err := client.Repositories.CreateStatus(ctx, params.Owner, params.Repo, params.Ref, status)

		if err != nil {
			log.Println("Error occurred while updating repo status on the project: ", err)
			return
		}
		log.Println("Successfully update project status to:", state)
	}
}

// Subscribe adds the sicuro webhook to the given repo
// TODO: save the access token of the user subscribing a repo to the database to use for offline actions such as build status update
func (client *GithubClient) Subscribe(params GithubRequestParams) error {
	active := true
	hook := github.Hook{
		Name:   github.String("web"),
		Active: &active,
		Events: []string{"push", "pull_request"},
		Config: map[string]interface{}{
			"content_type": "json",
			"url":          params.CallbackURL,
			"secret":       params.Creds,
		},
	}

	_, _, err := client.Repositories.CreateHook(ctx, params.Owner, params.Repo, &hook)
	if err != nil {
		log.Printf("Error %s occurred while creating webhook with params %v", err, params)

	}

	return err
}

// UserRepos returns a list of the github repos belongs to the user
// The user here refers to the owner of the access token used for the  github client
// Although the github api allows to explicitly specify a user,
// For now let it default to the user owning the access token
func (client *GithubClient) UserRepos() []*github.Repository {
	repos, _, err := client.Repositories.List(ctx, "", nil)
	if err != nil {
		log.Println("Error fetching users repo: ", err)
	}

	return repos
}

// Repo fetches and returns the github repo with the given params
func (client *GithubClient) Repo(params GithubRequestParams) (repo *github.Repository, err error) {
	repo, _, err = client.Repositories.Get(ctx, params.Owner, params.Repo)
	if err != nil {
		log.Printf("Error %s occurred fetching repo with params %v", err, params)
	}
	return
}

// IsRepoSubscribed checks if the given repo has the sicuro webhook set
func (client *GithubClient) IsRepoSubscribed(params GithubRequestParams) bool {
	ctx := ctx
	hooks, _, err := client.Repositories.ListHooks(ctx, params.Owner, params.Repo, &github.ListOptions{})

	if err != nil {
		log.Printf("Error %s occurred checking repo subscription status with params %v", err, params)
		return false
	}

	for _, hook := range hooks {
		if hasActiveWebhook(hook, params.CallbackURL) {
			return true
		}
	}

	return false
}

func hasActiveWebhook(hook *github.Hook, webhookPath string) bool {
	if !*(hook.Active) {
		return false
	}

	if hook.Config["url"] != webhookPath {
		return false
	}

	// TODO: also check that the webhook is subscribed to the right events => ["pull_request" "push"]
	return true
}
