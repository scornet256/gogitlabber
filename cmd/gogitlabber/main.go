package main

import (
	"fmt"

	"github.com/scornet256/go-logger"
)

// version
var version string
var config *Config

// repository data
type Repository struct {
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func main() {

	// set app version
	version = "2.2.0"

	// set appname for logger
	logger.SetAppName("gogitlabber")

	// manage all argument magic and load configuration
	config = manageArguments()

	// set debugging
	logger.SetDebug(config.Debug)

	// check for git
	err := verifyGitAvailable()
	if err != nil {
		logger.Fatal("VALIDATION: git not found in path", err)
	}
	logger.Print("VALIDATION: git found in path", nil)

	// validate git backend is set
	if config.GitBackend == "" {
		logger.Fatal("Configuration error: git_backend is required (gitlab|gitea)", nil)
	}

	// make initial progressbar
	if !config.Debug {
		progressBar()
	}

	// fetch repository information
	var repositories []Repository
	switch config.GitBackend {
	case "gitea":
		repositories, err = FetchRepositoriesGitea()
		if err != nil {
			logger.Fatal("Fetching repositories failed", err)
		}
	case "gitlab":
		repositories, err = FetchRepositoriesGitLab()
		if err != nil {
			logger.Fatal("Fetching repositories failed", err)
		}
	default:
		logger.Fatal(fmt.Sprintf("Unsupported git backend: %s (supported: gitlab|gitea)", config.GitBackend), nil)
	}

	// manage found repositories
	stats := &GitStats{}
	CheckoutRepositories(repositories, stats)
	printDetailedSummary(stats)
}
