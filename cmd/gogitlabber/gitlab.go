package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

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

	// make progressbar using:
	// - github.com/k0kubun/go-ansi
	// - github.com/schollz/progressbar/v3
	barPrefix := fmt.Sprintf("Getting your one and only repository...")
	if repoCount > 1 {
		barPrefix = fmt.Sprintf("Getting your repositories...")
	}

	bar := progressbar.NewOptions(
		repoCount,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetWidth(20),
		progressbar.OptionSetDescription(barPrefix),
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
