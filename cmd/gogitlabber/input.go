package main

import (
	"fmt"
	"os"
	"strings"
)

func loadEnvironmentVariables() error {
	gitlabHost = os.Getenv("GITLAB_HOSTNAME")
	gitlabToken = os.Getenv("GITLAB_API_TOKEN")
	repoDestinationPre = os.Getenv("GOGITLABBER_DESTINATION")
	return nil
}

func manageArguments() {

	// require at least the destination argument
	if len(os.Args) <= 1 {
		printUsage()
		os.Exit(1)
	}

	// parse arguments
	for _, arg := range os.Args[1:] {
		switch {

		case strings.HasPrefix(arg, "--archived="):
			includeArchived = strings.TrimPrefix(arg, "--archived=")

		case strings.HasPrefix(arg, "--destination="):
			repoDestinationPre = strings.TrimPrefix(arg, "--destination=")

		case strings.HasPrefix(arg, "--gitlab-api-token="):
			gitlabToken = strings.TrimPrefix(arg, "--gitlab-api-token=")

		case strings.HasPrefix(arg, "--gitlab-url="):
			gitlabHost = strings.TrimPrefix(arg, "--gitlab-url=")

		default:
			printUsage()
			os.Exit(1)
		}
	}

	// fail if destination is unknown
	if repoDestinationPre == "" {
		fmt.Println("Fatal: No destination found.")
		printUsage()
		os.Exit(1)
	}

	// add slash 🎩🎸 if not provided
	if !strings.HasSuffix(repoDestinationPre, "/") {
		repoDestinationPre += "/"
	}

	// --archive options:
	// - any      (fetch both)
	// - only     (fetch archived only)
	// - excluded (fetch non-archived only - default)
	if includeArchived == "" {
		includeArchived = "excluded"
	}

	if includeArchived != "any" &&
		includeArchived != "only" &&
		includeArchived != "excluded" {
		fmt.Println("Usage: gogitlabber --archived=(any|excluded|only)")
		os.Exit(1)
	}

	// verify GitLab input
	if gitlabHost == "" {
		fmt.Println("Fatal: No GitLab server configured.")
		printUsage()
		os.Exit(1)
	}

	if gitlabToken == "" {
		fmt.Println("Fatal: No GitLab API Token found.")
		printUsage()
		os.Exit(1)
	}
}
