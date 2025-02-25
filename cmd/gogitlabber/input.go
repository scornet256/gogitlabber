package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func manageArguments() {

	// configuration vars
  var archivedFlag = flag.String("archived", "excluded", "To include archived repositories (any|excluded|exclusive)\n  example: -archived=any\nenv = GOGITLABBER_ARCHIVED\n")
	var destinationFlag = flag.String("destination", "$HOME/Documents", "Specify where to check the repositories out\n  example: -destination=$HOME/repos\nenv = GOGITLABBER_DESTINATION\n")
	var hostFlag = flag.String("gitlab-url", "gitlab.com", "Specify GitLab host\n  example: -gitlab-url=gitlab.com\nenv = GITLAB_URL\n")
	var tokenFlag = flag.String("gitlab-api-token", "", "Specify GitLab API token\n  example: -gitlab-api=glpat-xxxx\nenv = GITLAB_API_TOKEN\n")

	flag.Parse()

	// assign the parsed values to your variables
	includeArchived = *archivedFlag
	repoDestinationPre = *destinationFlag
	gitlabToken = *tokenFlag
	gitlabHost = *hostFlag

	// use environment variable if set, otherwise use flag value
	if envHost := os.Getenv("GITLAB_URL"); envHost != "" {
		gitlabHost = envHost
	}

	if envToken := os.Getenv("GITLAB_API_TOKEN"); envToken != "" {
		gitlabToken = envToken
	}

	if envRepoDest := os.Getenv("GOGITLABBER_DESTINATION"); envRepoDest != "" {
		repoDestinationPre = envRepoDest
	}

	if envArchived := os.Getenv("GOGITLABBER_ARCHIVED"); envArchived != "" {
		includeArchived = envArchived
	}

	// fail if no configuration found
	if gitlabHost == "" {
		fmt.Println("Fatal: No GitLab Host found.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if gitlabToken == "" {
		fmt.Println("Fatal: No GitLab API Token found.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if repoDestinationPre == "" {
		fmt.Println("Fatal: No destination found.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// --archive options:
	// - any        (fetch both)
	// - exclusive  (fetch archived exclusive)
	// - excluded   (fetch non-archived exclusive - default)
	if includeArchived == "" {
		includeArchived = "excluded"
	}

	if includeArchived != "any" &&
		includeArchived != "exclusive" &&
		includeArchived != "excluded" {
		fmt.Println("Fatal: Wrong archive option found.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// add slash 🎩🎸 if not provided
	if !strings.HasSuffix(repoDestinationPre, "/") {
		repoDestinationPre += "/"
	}

}
