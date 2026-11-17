package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/scornet256/go-logger"
)

// gitea api client
type GiteaClient struct {
	httpClient *http.Client
	apiBase    string
	token      string
}

// gitea repository as returned by the api
type giteaRepository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Archived bool   `json:"archived"`
}

func NewGiteaClient(host, token string) *GiteaClient {
	return &GiteaClient{
		httpClient: newHTTPClient(),
		apiBase:    "https://" + host + "/api/v1",
		token:      token,
	}
}

// fetch all repositories of the user, following pagination
func (c *GiteaClient) FetchRepositories(ctx context.Context, includeArchived string) ([]Repository, error) {
	var repositories []Repository
	headers := map[string]string{"Authorization": "token " + c.token}

	for page := 1; ; page++ {
		var repos []giteaRepository
		respHeaders, err := getJSON(ctx, c.httpClient, c.reposURL(page), headers, &repos)
		if err != nil {
			return nil, fmt.Errorf("fetching page %d: %w", page, err)
		}

		// /user/repos has no archived filter, so filter here
		for _, repo := range repos {
			if keepArchived(repo.Archived, includeArchived) {
				repositories = append(repositories, Repository{
					Name:              repo.Name,
					PathWithNamespace: repo.FullName,
				})
			}
		}

		logger.Print(fmt.Sprintf("Fetched page %d (%d repositories)", page, len(repos)), nil)

		if len(repos) == 0 || !strings.Contains(respHeaders.Get("Link"), `rel="next"`) {
			break
		}
	}

	return repositories, nil
}

// build repos api url; gitea caps limit at its MAX_RESPONSE_ITEMS setting
func (c *GiteaClient) reposURL(page int) string {
	query := url.Values{}
	query.Set("limit", "50")
	query.Set("page", strconv.Itoa(page))

	return c.apiBase + "/user/repos?" + query.Encode()
}
