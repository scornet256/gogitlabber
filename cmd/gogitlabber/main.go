package main

import (
	"log"
)

// userdata
var repoDestinationPre string
var includeArchived string
var gitlabToken string
var gitlabHost string

// functional vars
var pullError []string

type Repository struct {
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func main() {

	// manage all argument magic
	manageArguments()

	// fetch repository information from gitlab
	repositories, err := fetchRepositories()
	if err != nil {
		log.Fatalf("Error fetching repositories: %v", err)
	}

	// manage found repositories
	checkoutRepositories(repositories)
	printPullerror(pullError)
}
