package main

import (
	"github.com/scornet256/go-logger"
)

// version
var version string

// userdata
var concurrency int
var debug bool
var includeArchived string
var repoDestinationPre string

// git
var gitHost string
var gitToken string
var gitBackend string

// keep count 🧛
var clonedCount int
var errorCount int
var pulledCount int
var pullErrorMsgUnstaged []string
var pullErrorMsgUncommitted []string

// repository data
type Repository struct {
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func main() {

	// set app version
	version = "1.0.0"

	// set appname for logger
	logger.SetAppName("gogitlabber")

	// manage all argument magic
	manageArguments()

	// set debugging
	logger.SetDebug(debug)

	// check for git
	err := verifyGitAvailable()
	if err != nil {
		logger.Fatal("VALIDATION: git not found in path", err)
	}
	logger.Print("VALIDATION: git found in path", nil)

	// make initial progressbar
	if !debug {
		progressBar()
	}

	// fetch repository information
	var repositories []Repository
	switch gitBackend {
	case "gitea":
		repositories, err = fetchRepositoriesGitea()
		if err != nil {
			logger.Fatal("Fetching repositories failed", err)
		}
	case "gitlab":
		repositories, err = fetchRepositoriesGitlab()
		if err != nil {
			logger.Fatal("Fetching repositories failed", err)
		}
	default:
		logger.Fatal("Fetching repositories failed", err)
	}

	// manage found repositories
	checkoutRepositories(repositories, concurrency)
	printSummary()
	printPullErrorUnstaged(pullErrorMsgUnstaged)
	printPullErrorUncommitted(pullErrorMsgUncommitted)
}
