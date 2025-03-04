package main

import (
	"flag"
	"log"
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
		flag.Usage()
    log.Printf("fatal: config; gitlab api token not found")
	}

	// manage gitlab url option
	switch envHost := os.Getenv("GITLAB_URL"); {
	case envHost != "":
		gitlabHost = envHost
	default:
		flag.Usage()
    log.Fatalf("fatal: config; gitlab host not found")
	}

	// manage destination option
	switch envRepoDest := os.Getenv("GOGITLABBER_DESTINATION"); {
	case envRepoDest != "":
		repoDestinationPre = envRepoDest
	default:
		flag.Usage()
    log.Fatalf("fatal: config; destination not found")
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
		flag.Usage()
    log.Fatalf("fatal: config; no or wrong archive option found")
	}
}
