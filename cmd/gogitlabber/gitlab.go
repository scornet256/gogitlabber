package main

import (
	"encoding/json"
	"fmt"
	"gogitlabber/cmd/gogitlabber/logging"
	"net/http"
)

func fetchRepositoriesGitlab() ([]Repository, error) {

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

	logging.Print("HTTP: Creating API request", nil)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ERROR: creating request: %v\n", err)
	}

	logging.Print("HTTP: Adding PRIVATE-TOKEN header to API request", nil)
	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	logging.Print("HTTP: Making request", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ERROR: making request: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ERROR: API request failed with status: %d\n", resp.StatusCode)
	}

	logging.Print("HTTP: Decoding JSON response", nil)
	var repositories []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
		return nil, fmt.Errorf("ERROR: decoding response: %v\n", err)
	}

	if len(repositories) < 1 {
		return repositories, fmt.Errorf("ERROR: no repositories found\n")
	}

	repoCount := len(repositories)

	logging.Print("Resetting the progressbar", nil)
	if !debug {
		err = bar.Set(0)
		if err != nil {
			logging.Fatal("Could not reset the progressbar", err)
		}
	}

	logging.Print("Increasing the max value of the progressbar", nil)
	if !debug {
		bar.ChangeMax(repoCount)
	}

	logging.Print("HTTP: Returning repositories found", nil)
	return repositories, nil
}
