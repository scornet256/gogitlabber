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

	// make progressbar
	barPrefix := fmt.Sprintf("Getting your one and only repository...")
	if repoCount > 1 {
		barPrefix = fmt.Sprintf("Getting your repositories...")
	}

	bar := progressbar.NewOptions(
		repoCount,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowDescriptionAtLineEnd(),
		progressbar.OptionSetElapsedTime(false),
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

		descriptionPrefixPre := "Cloning repository "
		descriptionPrefix := descriptionPrefixPre + repoName
		bar.Describe(descriptionPrefix)

		cloneOutput, err := cloneRepository(repoDestination, gitlabUrl)

		if err != nil {

			// if repo already exists, try to pull the latest changes
			if strings.Contains(string(cloneOutput),
				"already exists and is not an empty directory") {

				descriptionPrefixPre := "Pulling repository "
				descriptionPrefix := descriptionPrefixPre + repoName
				bar.Describe(descriptionPrefix)

				_, err := pullRepositories(repoDestination)
				if err != nil {
					bar.Add(1)
					continue
				}

				pulledCount = pulledCount + 1
				bar.Add(1)
				continue
			}

			log.Printf("\n❌ error cloning %s: %v\n%s\n", repoName, err, cloneOutput)
			errorCount = errorCount + 1
			bar.Add(1)
			continue
		}

		clonedCount = clonedCount + 1
		bar.Add(1)
	}

	// print empty line as the bar does not do that
	fmt.Println("")

	// print summary
	fmt.Printf(
		"Summary:\n"+
			" Cloned repositories: %v\n"+
			" Pulled repositories: %v\n"+
			" Errors: %v\n",
		clonedCount,
		pulledCount,
		errorCount,
	)
}

func cloneRepository(repoDestination string, gitlabUrl string) (string, error) {

	cloneCmd := exec.Command("git", "clone", gitlabUrl, repoDestination)
	cloneOutput, err := cloneCmd.CombinedOutput()

	return string(cloneOutput), err
}

func findRemote(repoDestination string) (string, error) {

	remoteCmd := exec.Command("git", "-C", repoDestination, "remote", "show")
	remoteOutput, err := remoteCmd.CombinedOutput()
	remote := strings.Split(strings.TrimSpace(string(remoteOutput)), "\n")[0]

	if err != nil {
		log.Printf("\n❌ error finding remote for: %s\n", err)
	}

	return remote, err
}

func pullRepositories(repoDestination string) (string, error) {

	remote, err := findRemote(repoDestination)
	pullCmd := exec.Command("git", "-C", repoDestination, "pull", remote)
	pullOutput, err := pullCmd.CombinedOutput()

	if err != nil {
		errorCount = errorCount + 1
		if strings.Contains(string(pullOutput), "You have unstaged changes") {
			pullErrorMsg = append(pullErrorMsg, repoDestination)
		}
	}

	return string(pullOutput), err
}

func printPullError(pullErrorMsg []string) {
	if len(pullErrorMsg) > 0 {
		for _, repo := range pullErrorMsg {
			fmt.Printf("❕%s has unstaged changes.\n", repo)
		}
	}
}
