package main

import (
	"encoding/json"
	"fmt"
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

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ERROR: creating request: %v\n", err)
	}

	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ERROR: making request: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ERROR: API request failed with status: %d\n", resp.StatusCode)
	}

	var repositories []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
		return nil, fmt.Errorf("ERROR: decoding response: %v\n", err)
	}

	if len(repositories) < 1 {
		return repositories, fmt.Errorf("ERROR: no repositories found\n")
	}

	return repositories, nil
}
