package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func manageArguments() {

	var archivedFlag = flag.String("archived", "excluded", "to include archived repositories (any|excluded|exclusive)\nenv = GOGITLABBER_ARCHIVED\n")
	var destinationFlag = flag.String("destination", "", "where to check the repositories out\nenv = GOGITLABBER_DESTINATION")
	var tokenFlag = flag.String("gitlab-api-token", "", "gitlab api token; example glpat-xxxx\nenv = GITLAB_API_TOKEN")
	var hostFlag = flag.String("gitlab-url", "", "gitlab host; example gitlab.example.com\nenv = GITLAB_HOSTNAME")

	flag.Parse()

	// assign the parsed values to your variables
	includeArchived = *archivedFlag
	repoDestinationPre = *destinationFlag
	gitlabToken = *tokenFlag
	gitlabHost = *hostFlag

	// use environment variable if set, otherwise use flag value
	if envHost := os.Getenv("GITLAB_HOSTNAME"); envHost != "" {
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
