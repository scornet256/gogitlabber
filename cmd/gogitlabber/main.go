package main

// userdata
var repoDestinationPre string
var includeArchived string
var gitlabToken string
var gitlabHost string

// keep count 🧛
var clonedCount int
var errorCount int
var pulledCount int
var pullErrorMsg []string

// repository data
type Repository struct {
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func main() {

	// manage all argument magic
	manageArguments()

	// check for git
	verifyGitAvailable()

	// fetch repository information from gitlab
	repositories := fetchRepositoriesGitlab()

	// manage found repositories
	progressBar(repositories)
	checkoutRepositories(repositories)
	printSummary()
	printPullError(pullErrorMsg)
}
