package main

import (
	"encoding/json"
	"fmt"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// userdata
var repoDestinationPre string
var includeArchived string
var gitlabToken string
var gitlabHost string

// functional vars
var pullError []string

type Repository struct {
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func main() {

	// environment variables < arguments
	if err := loadEnvironmentVariables(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	// manage all argument magic
	manageArguments()

	// fetch repository information from gitlab
	repositories, err := fetchRepositories()
	if err != nil {
		log.Fatalf("Error fetching repositories: %v", err)
	}

	// manage found repositories
	checkoutRepositories(repositories)
	printPullerror(pullError)
}

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

func fetchRepositories() ([]Repository, error) {

	// default options
	membership := "membership=true"
	perpage := "per_page=100"
	order := "order_by=name"

	// configure archived options
	var archived string
	switch {
	case includeArchived == "excluded":
		archived = "&archived=false"
	case includeArchived == "only":
		archived = "&archived=true"
	default:
		archived = ""
	}

	url := fmt.Sprintf("https://%s/api/v4/projects?%s&%s&%s%s",
		gitlabHost, membership, order, perpage, archived)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var repositories []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return repositories, nil
}

func checkoutRepositories(repositories []Repository) {
	repoCount := len(repositories)

	fmt.Printf("Found %d repositories", repoCount)

	// make progressbar using:
	// - github.com/k0kubun/go-ansi
	// - github.com/schollz/progressbar/v3
	bar := progressbar.NewOptions(
		repoCount,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription("Getting your repositories..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	for _, repo := range repositories {

		repoName := string(repo.PathWithNamespace)
		gitlabUrl := fmt.Sprintf("https://gitlab-token:%s@%s/%s.git",
			gitlabToken, gitlabHost, repoName)

		repoDestination := repoDestinationPre + repoName

		cloneCmd := exec.Command("git", "clone", gitlabUrl, repoDestination)
		cloneOutput, err := cloneCmd.CombinedOutput()

		if err != nil {

			// if repo already exists, try to pull the latest changes
			if strings.Contains(string(cloneOutput),
				"already exists and is not an empty directory") {
				pullRepositories(repoDestination)
				bar.Add(1)
				continue
			}
			log.Printf("❌ Error cloning %s: %v\n%s", repoName, err, string(cloneOutput))
			bar.Add(1)
			continue
		}
		bar.Add(1)
	}

	// print empty line as the bar does not do that
	fmt.Println("")
}

func pullRepositories(repoDestination string) {
	pullCmd := exec.Command("git", "-C", repoDestination, "pull", "origin")
	pullOutput, err := pullCmd.CombinedOutput()

	if err != nil {
		if strings.Contains(string(pullOutput),
			"You have unstaged changes") {
			pullError = append(pullError, repoDestination)
		}
	}
}

func printPullerror(pullError []string) {
	if len(pullError) > 0 {
		for _, repo := range pullError {
			fmt.Printf("❕%s has unstaged changes.\n", repo)
		}
	}
}

func printUsage() {
	fmt.Println("Usage: gogitlabber")
	fmt.Println("         --archived=(any|excluded|only)")
	fmt.Println("         --destination=$HOME/Documents")
	fmt.Println("         --gitlab-url=gitlab.example.com")
	fmt.Println("         --gitlab-token=<supersecrettoken>")
	fmt.Println("")
	fmt.Println("You can also set these environment variables:")
	fmt.Println("  GOGITLABBER_ARCHIVED=(any|excluded|only)")
	fmt.Println("  GOGITLABBER_DESTINATION=$HOME/Documents")
	fmt.Println("  GITLAB_API_TOKEN=<supersecrettoken>")
	fmt.Println("  GITLAB_URL=gitlab.example.com")
	fmt.Println("")
}
