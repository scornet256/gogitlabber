package main

import (
	"fmt"
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

	// environment variables < arguments
	if err := loadEnvironmentVariables(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

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

func printUsage() {
	fmt.Println("Usage: gogitlabber")
	fmt.Println("         --archived=(any|excluded|only)")
	fmt.Println("         --destination=$HOME/Documents")
	fmt.Println("         --gitlab-url=gitlab.example.com")
	fmt.Println("         --gitlab-token=<supersecrettoken>")
	fmt.Println("")
	fmt.Println("You can also set these environment variables:")
	fmt.Println("  GOGITLABBER_ARCHIVED=(any|excluded|only)")
	fmt.Println("  GOGITLABBER_DESTINATION=$HOME/Documents")
	fmt.Println("  GITLAB_API_TOKEN=<supersecrettoken>")
	fmt.Println("  GITLAB_URL=gitlab.example.com")
	fmt.Println("")
}
