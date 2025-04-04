package main

import (
	"github.com/scornet256/go-logger"
)

// version
var version string

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

	// set app version
	version = "0.0.9"

	// set appname for logger
	logger.SetAppName("gogitlabber")

	// manage all argument magic
	manageArguments()

	// set debugging
	logger.SetDebug(debug)

	// check for git
	err := verifyGitAvailable()
	if err != nil {
		logger.Fatal("VALIDATION: git not found in path", err)
	}
	logger.Print("VALIDATION: git found in path", nil)

	// make initial progressbar
	if !debug {
		progressBar()
	}

	// fetch repository information from gitlab
	repositories, err := fetchRepositoriesGitlab()
	if err != nil {
		logger.Fatal("Fetching repositories failed", err)
	}

	// manage found repositories
	checkoutRepositories(repositories, concurrency)
	printSummary()
	printPullError(pullErrorMsg)
}
