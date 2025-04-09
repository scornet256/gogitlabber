package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/scornet256/go-logger"
)

func fetchRepositoriesGitea() ([]Repository, error) {

	type GiteaRepository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	}

	// default options
	visibility := "visibility=all"
	perpage := "limit=100"
	sort := "sort=alpha"

	// configure archived options
	var archived string
	switch includeArchived {
	case "excluded":
		archived = "&archived=false"
	case "only":
		archived = "&archived=true"
	default:
		archived = ""
	}

	url := fmt.Sprintf("https://%s/api/v1/user/repos?%s&%s&%s%s",
		gitHost, visibility, sort, perpage, archived)

	logger.Print("HTTP: Creating API request", nil)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ERROR: creating request: %v", err)
	}

	logger.Print("HTTP: Adding Authorization header to API request", nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", gitToken))

	logger.Print("HTTP: Making request", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ERROR: making request: %v", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Fatal("HTTP: Error closing response body", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ERROR: API request failed with status: %d", resp.StatusCode)
	}
	logger.Print("HTTP: Decoding JSON response", nil)

	// first decode into gitearepository slice
	var giteaRepos []GiteaRepository
	if err := json.NewDecoder(resp.Body).Decode(&giteaRepos); err != nil {
		return nil, fmt.Errorf("ERROR: decoding response: %v", err)
	}

	// convert to repository slice
	repositories := make([]Repository, len(giteaRepos))
	for repo, giteaRepo := range giteaRepos {
		repositories[repo] = Repository{
			Name:              giteaRepo.Name,
			PathWithNamespace: giteaRepo.FullName,
		}
	}

	if len(repositories) < 1 {
		return repositories, fmt.Errorf("ERROR: no repositories found")
	}
	repoCount := len(repositories)

	logger.Print("BAR: Resetting the progressbar", nil)
	if !debug {
		err = bar.Set(0)
		if err != nil {
			logger.Fatal("Could not reset the progressbar", err)
		}
	}

	logger.Print("BAR: Increasing the max value of the progressbar", nil)
	if !debug {
		bar.ChangeMax(repoCount)
	}

	logger.Print("HTTP: Returning repositories found", nil)
	return repositories, nil
}
