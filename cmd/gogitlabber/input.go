package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

func manageArguments() {

	// configuration vars
	var archivedFlag = flag.String("archived", "excluded", "To include archived repositories (any|excluded|exclusive)\n  example: -archived=any\nenv = GOGITLABBER_ARCHIVED\n")
	var concurrencyFlag = flag.Int("concurrency", 15, "Specify repository concurrency\n  example: -concurrency=15\nenv = GOGITLABBER_CONCURRENCY\n")
	var destinationFlag = flag.String("destination", "$HOME/Documents", "Specify where to check the repositories out\n  example: -destination=$HOME/repos\nenv = GOGITLABBER_DESTINATION\n")
	var hostFlag = flag.String("gitlab-url", "gitlab.com", "Specify GitLab host\n  example: -gitlab-url=gitlab.com\nenv = GITLAB_URL\n")
	var tokenFlag = flag.String("gitlab-api-token", "", "Specify GitLab API token\n  example: -gitlab-api=glpat-xxxx\nenv = GITLAB_API_TOKEN\n")

	flag.Parse()

	// assign the parsed values to your variables
	concurrency = *concurrencyFlag
	gitlabHost = *hostFlag
	gitlabToken = *tokenFlag
	includeArchived = *archivedFlag
	repoDestinationPre = *destinationFlag

	// manage gitlab api option
	switch envToken := os.Getenv("GITLAB_API_TOKEN"); {
	case envToken != "":
		gitlabToken = envToken
	default:
		flag.Usage()
		log.Printf("FATAL: config; gitlab api token not found\n")
	}

	// manage gitlab url option
	switch envHost := os.Getenv("GITLAB_URL"); {
	case envHost != "":
		gitlabHost = envHost
	default:
		flag.Usage()
		log.Fatalf("FATAL: config; gitlab host not found\n")
	}

	// manage destination option
	switch envRepoDest := os.Getenv("GOGITLABBER_DESTINATION"); {
	case envRepoDest != "":
		repoDestinationPre = envRepoDest
	default:
		flag.Usage()
		log.Fatalf("FATAL: config; destination not found\n")
	}

	// add slash 🎩🎸 if not provided
	switch {
	case !strings.HasSuffix(repoDestinationPre, "/"):
		repoDestinationPre += "/"
	}

	// manage concurrency option
  switch envConcurrency := os.Getenv("GOGITLABBER_CONCURRENCY"); {
  case envConcurrency == "":
    concurrency = 15
  case envConcurrency != "":
		concurrencyValue, err := strconv.Atoi(envConcurrency)
		if err != nil {
      log.Fatalf("FATAL: invalid concurrency value in environment: %v", err)
		}
		concurrency = concurrencyValue
  default:
    flag.Usage()
    log.Fatalf("FATAL: config; concurrency not found\n")
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
		log.Fatalf("FATAL: config; no or wrong archive option found\n")
	}
}
