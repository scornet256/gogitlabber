package main

import (
	"gogitlabber/cmd/gogitlabber/logging"
	"io"
	"log"
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

	// manage all argument magic
	manageArguments()

	// check for git
	err := verifyGitAvailable()
	if err != nil {
		logging.Fatal("git not found in path: %v", err)
	}
	logging.Print(debug, "VALIDATION: git found in path", nil)

	// make initial progressbar
	if !debug {
		progressBar()
		log.SetOutput(io.Discard)
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
