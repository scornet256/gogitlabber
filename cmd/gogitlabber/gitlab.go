package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	for _, repo := range repositories {

		// create clone gitlab url
		repoName := string(repo.PathWithNamespace)
		gitlabUrl := fmt.Sprintf("https://gitlab-token:%s@%s/%s.git",
			gitlabToken, gitlabHost, repoName)

		// create repository destination
		repoDestination := repoDestinationPre + repoName

		// create bar description
		descriptionPrefixPre := "Cloning repository "
		descriptionPrefix := descriptionPrefixPre + repoName + " ..."
		bar.Describe(descriptionPrefix)

		// clone the repo
		cloneOutput, err := cloneRepository(repoDestination, gitlabUrl)

		if err != nil {

			// if repo already exists, try to pull the latest changes
			if strings.Contains(string(cloneOutput),
				"already exists and is not an empty directory") {

				descriptionPrefixPre := "Pulling repository "
				descriptionPrefix := descriptionPrefixPre + repoName + " ..."
				bar.Describe(descriptionPrefix)

				_, err := pullRepositories(repoDestination)
				if err != nil {
					continue
				}

				pulledCount = pulledCount + 1
				continue
			}

			// in case cloning failed and the directory does not exist
			// print the clone error and continue
			log.Printf("\n❌ error cloning %s: %v\n%s\n", repoName, err, cloneOutput)
			errorCount = errorCount + 1
			bar.Add(1)
			continue
		}

		// finish the clone
		clonedCount = clonedCount + 1
		bar.Add(1)
	}
}
