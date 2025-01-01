package main

import (
	"encoding/json"
	"fmt"
	"log"
  "strings"
	"net/http"
	"os"
	"os/exec"
  "github.com/k0kubun/go-ansi"
  "github.com/schollz/progressbar/v3"
)

type Repository struct {
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

var repoDestinationPre string
var includeArchived string
var gitlabToken string
var	gitlabHost  string

func main() {

  // load environment variables first, they will be overridden
  // by argument flags if specified.
	if err := loadEnvironmentVariables(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

  // require at least the destination argument
  if len(os.Args) <= 1 {
		fmt.Println("Usage:   gogitlabber --destination=<directory>")
		fmt.Println("Example: gogitlabber --destination=/tmp/repos")
		os.Exit(1)
	}

  // parse arguments
	for _, arg := range os.Args[1:] {
    switch {

    case strings.HasPrefix(arg, "--destination="):
			repoDestinationPre = strings.TrimPrefix(arg, "--destination=")

    case strings.HasPrefix(arg, "--gitlab-api-token="):
      gitlabToken = strings.TrimPrefix(arg, "--gitlab-api-token=")

    case strings.HasPrefix(arg, "--gitlab-url="):
			gitlabHost = strings.TrimPrefix(arg, "--gitlab-url=")

    default:
      fmt.Println("Usage:   gogitlabber --destination=<directory>")
      fmt.Println("Example: gogitlabber --destination=/tmp/repos")
      os.Exit(1)
    }
	}

  // fail if destination is unknown
  if repoDestinationPre == "" {
    fmt.Println("Fatal: No destination found.")
		fmt.Println("Example: gogitlabber --destination=/tmp/repos")
		fmt.Println("Usage: gogitlabber --destination=/tmp/repos")
    os.Exit(1)
  }

  // fetch repository information
	repositories, err := fetchRepositories()
	if err != nil {
		log.Fatalf("Error fetching repositories: %v", err)
	}

  // manage found repositories
	checkoutRepositories(repositories)
}


func loadEnvironmentVariables() error {
	gitlabToken = os.Getenv("GITLAB_API_KEY")
	if gitlabToken == "" {
		return fmt.Errorf("GITLAB_API_KEY environment variable is not set")
	}

	gitlabHost = os.Getenv("GITLAB_HOSTNAME")
	if gitlabHost == "" {
		return fmt.Errorf("GITLAB_HOSTNAME environment variable is not set")
	}
	return nil
}


func fetchRepositories() ([]Repository, error) {

  archived   := "archived=false"
  membership := "membership=true"
  perpage    := "per_page=100"
  order      := "order_by=name"

	url := fmt.Sprintf("https://%s/api/v4/projects?%s&%s&%s&%s", 
                     gitlabHost, membership, order, archived, perpage)

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

		cloneCmd         := exec.Command("git", "clone", gitlabUrl, repoDestination)
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
}


func pullRepositories(repoDestination string) {
	pullCmd     := exec.Command("git", "-C", repoDestination, "pull", "origin")
  output, err := pullCmd.CombinedOutput()

  if err != nil {
    log.Printf("❌ Error pulling %s: %v\n%s", repoDestination, err, string(output))
  }
}
