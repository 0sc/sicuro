package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/context"
)

const (
	sessionName = "sicuro-auth"
)

var (
	port   = os.Getenv("PORT")
	appDIR = filepath.Join(os.Getenv("ROOT_DIR"), "app")
)

func main() {
	setupGithubOAuth()
	registerRoutes()

	fmt.Printf("Starting server on port: %s\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
