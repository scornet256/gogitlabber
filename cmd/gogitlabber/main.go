package main

import (
	"io"
	"log"
)

// userdata
var concurrency int
var gitlabHost string
var gitlabToken string
var includeArchived string
var repoDestinationPre string
var verbose bool

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
    logFatal("FATAL: git not found in path: %v", err)
	}
  logPrint("Git is available. Proceeding with the program.", nil)

	// fetch repository information from gitlab
	repositories, err := fetchRepositoriesGitlab()
	if err != nil {
		logFatal("FATAL: %v", err)
	}
  logPrint("Logged into GitLab, Repositories found. Proceeding with the program.", nil)

  // print progressbar ony if not in verbose mode
  if !verbose {
	  progressBar(repositories)
    log.SetOutput(io.Discard)
  }

	// manage found repositories
	checkoutRepositories(repositories, concurrency)
	printSummary()
	printPullError(pullErrorMsg)
}
