package main

import (
	"gogitlabber/cmd/gogitlabber/logging"
)

// userdata
var concurrency int
var debug bool
var gitlabHost string
var gitlabToken string
var includeArchived string
var repoDestinationPre string

// keep count 🧛
var clonedCount int
var errorCount int
var pulledCount int
var pullErrorMsg []string

// repository data
type Repository struct {
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func main() {

	// set appname for logging
	logging.SetAppName("gogitlabber")

	// manage all argument magic
	manageArguments()

	// set debugging
	logging.SetDebug(debug)

	// check for git
	err := verifyGitAvailable()
	if err != nil {
		logging.Fatal("git not found in path: %v", err)
	}
	logging.Print("VALIDATION: git found in path", nil)

	// make initial progressbar
	if !debug {
		progressBar()
	}

	// fetch repository information from gitlab
	repositories, err := fetchRepositoriesGitlab()
	if err != nil {
		logging.Fatal("FATAL: %v", err)
	}

	// manage found repositories
	checkoutRepositories(repositories, concurrency)
	printSummary()
	printPullError(pullErrorMsg)
}
