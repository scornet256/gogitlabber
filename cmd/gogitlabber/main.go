package main

import (
	"fmt"

	"github.com/scornet256/go-logger"
)

// version
var version string
var globalConfig *Config

// repository data
type Repository struct {
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func main() {

	// set app version
	version = "3.0.1"

	// set appname for logger
	logger.SetAppName("gogitlabber")

	// manage all argument magic and load configuration
	globalConfig = manageArguments()

	// set debugging
	logger.SetDebug(globalConfig.Debug)

	// validate git backend is set
	if globalConfig.GitBackend == "" {
		logger.Fatal("Configuration error: git_backend is required (gitlab|gitea)", nil)
	}

	// make initial progressbar
	if !globalConfig.Debug {
		progressBar()
	}

	// fetch repository information
	var repositories []Repository
	var err error
	switch globalConfig.GitBackend {
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
		logger.Fatal(fmt.Sprintf("Unsupported git backend: %s (supported: gitlab|gitea)", globalConfig.GitBackend), nil)
	}

	// manage found repositories
	stats := &GitStats{}
	CheckoutRepositories(repositories, stats)
	printDetailedSummary(stats)
}
