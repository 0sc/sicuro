package main

import "net/http"

const (
	runCIPath       = "/run"
	showPath        = "/show"
	indexPath       = "/index"
	dashboardPath   = "/dashboard"
	ciPath          = "/ci/"
	ghAuthPath      = "/gh/auth"
	ghSubscribePath = "/gh/subscribe"
	ghCallbackPath  = "/gh/callback"
	ghWebhookPath   = "/gh/webhook"
	websocketPath   = "/ws/"
)

func registerRoutes() {
	http.HandleFunc(ciPath, ciPageHandler())
	http.HandleFunc(runCIPath, runCIHandler())
	http.HandleFunc(showPath, showPageHandler())
	http.HandleFunc(indexPath, indexPageHandler())
	http.HandleFunc(dashboardPath, dashboardPageHandler())
	http.HandleFunc(ghSubscribePath, githubSubscriptionHandler())

	http.HandleFunc(websocketPath, wsHandler)

	http.HandleFunc(ghAuthPath, ghAuth)
	http.HandleFunc(ghCallbackPath, ghAuthCallback)
	http.HandleFunc(ghWebhookPath, githubWebhookHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, indexPath, http.StatusPermanentRedirect)
	})
}
