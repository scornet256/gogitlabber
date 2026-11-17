package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/scornet256/go-logger"
)

// minimum access level: 10=Guest, 20=Reporter, 30=Developer, 40=Maintainer, 50=Owner
const gitLabMinAccessLevel = 20

// gitlab api client
type GitLabClient struct {
	httpClient *http.Client
	apiBase    string
	token      string
}

// gitlab project as returned by the api
type gitLabProject struct {
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	Archived          bool   `json:"archived"`
}

func NewGitLabClient(host, token string) *GitLabClient {
	return &GitLabClient{
		httpClient: newHTTPClient(),
		apiBase:    "https://" + host + "/api/v4",
		token:      token,
	}
}

// fetch all projects the user is a member of, following pagination
func (c *GitLabClient) FetchRepositories(ctx context.Context, includeArchived string) ([]Repository, error) {
	var repositories []Repository
	headers := map[string]string{"PRIVATE-TOKEN": c.token}

	for page := 1; page > 0; {
		var projects []gitLabProject
		respHeaders, err := getJSON(ctx, c.httpClient, c.projectsURL(page, includeArchived), headers, &projects)
		if err != nil {
			return nil, fmt.Errorf("fetching page %d: %w", page, err)
		}

		for _, project := range projects {
			if keepArchived(project.Archived, includeArchived) {
				repositories = append(repositories, Repository{
					Name:              project.Name,
					PathWithNamespace: project.PathWithNamespace,
				})
			}
		}

		logger.Print(fmt.Sprintf("Fetched page %d (%d projects)", page, len(projects)), nil)

		// X-Total-Pages is omitted for large result sets, so only rely on X-Next-Page
		next, _ := strconv.Atoi(respHeaders.Get("X-Next-Page"))
		if next <= page {
			break
		}
		page = next
	}

	return repositories, nil
}

// build projects api url
func (c *GitLabClient) projectsURL(page int, includeArchived string) string {
	query := url.Values{}
	query.Set("membership", "true")
	query.Set("min_access_level", strconv.Itoa(gitLabMinAccessLevel))
	query.Set("order_by", "name")
	query.Set("sort", "asc")
	query.Set("per_page", "100")
	query.Set("page", strconv.Itoa(page))
	if archived := archivedParam(includeArchived); archived != "" {
		query.Set("archived", archived)
	}

	return c.apiBase + "/projects?" + query.Encode()
}
