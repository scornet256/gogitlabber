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

  // manage gitlab api option
	switch envToken := os.Getenv("GITLAB_API_TOKEN"); {
	case envToken != "":
		gitlabToken = envToken
	default:
		fmt.Println("Fatal: No GitLab API Token found.")
		flag.PrintDefaults()
		os.Exit(1)
	}

  // manage gitlab url option
	switch envHost := os.Getenv("GITLAB_URL"); {
	case envHost != "":
		gitlabHost = envHost
	default:
		fmt.Println("Fatal: No GitLab Host found.")
		flag.PrintDefaults()
		os.Exit(1)
	}

  // manage destination option
	switch envRepoDest := os.Getenv("GOGITLABBER_DESTINATION"); {
	case envRepoDest != "":
		repoDestinationPre = envRepoDest
	default:
		fmt.Println("Fatal: No destination found.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// add slash 🎩🎸 if not provided
  switch {
  case !strings.HasSuffix(repoDestinationPre, "/"):
		repoDestinationPre += "/"
	}

  // manage archived option
	switch envArchived := os.Getenv("GOGITLABBER_ARCHIVED"); {
  case envArchived == "":
		includeArchived = "excluded"

	case envArchived == "any":
		includeArchived = envArchived

	case envArchived == "exclusive":
		includeArchived = envArchived

	case envArchived == "excluded":
		includeArchived = envArchived

  default:
		fmt.Println("Fatal: Wrong archive option found.")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
