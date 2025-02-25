package main

import (
	"log"
)

// userdata
var repoDestinationPre string
var includeArchived string
var gitlabToken string
var gitlabHost string

var clonedCount int
var errorCount int
var pulledCount int
var pullErrorMsg []string

type Repository struct {
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func main() {

	// manage all argument magic
	manageArguments()

  // check for git
  verifyGitAvailable()

	// fetch repository information from gitlab
	repositories, err := fetchRepositoriesGitlab()
	if err != nil {
		log.Fatalf("Error fetching repositories: %v", err)
	}

	// manage found repositories
	progressBar(repositories)
	checkoutRepositories(repositories)
	printSummary()
	printPullError(pullErrorMsg)
}
