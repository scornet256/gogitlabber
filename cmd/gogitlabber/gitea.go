package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/scornet256/go-logger"
)

// giteaClient struct
type GiteaClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// gitea repo information
type GiteaRepository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

// gitea api options
type GiteaAPIOptions struct {
	Visibility      string
	IncludeArchived string
	Sort            string
	Limit           int
	Page            int
}

// gitea api client
func NewGiteaClient(baseURL, token string) *GiteaClient {
	return &GiteaClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
		token:   token,
	}
}

// fetch gitea repos
func FetchRepositoriesGitea() ([]Repository, error) {
	client := NewGiteaClient(globalConfig.GitHost, globalConfig.GitToken)
	options := GiteaAPIOptions{
		Visibility:      "all",
		IncludeArchived: globalConfig.IncludeArchived,
		Sort:            "alpha",
		Limit:           100,
		Page:            1,
	}

	repositories, err := client.fetchAllRepositories(context.Background(), options)
	if err != nil {
		return nil, fmt.Errorf("fetching repositories: %w", err)
	}

	if len(repositories) == 0 {
		return repositories, fmt.Errorf("no repositories found")
	}

	// update progress bar
	if err := updateProgressBar(len(repositories)); err != nil {
		logger.Print("WARNING: failed to update progress bar: "+err.Error(), nil)
	}

	logger.Print(fmt.Sprintf("Successfully fetched %d repositories", len(repositories)), nil)
	return repositories, nil
}

// fetch all repos with pagination
func (c *GiteaClient) fetchAllRepositories(ctx context.Context, options GiteaAPIOptions) ([]Repository, error) {
	var allRepositories []Repository

	for {
		giteaRepos, hasMore, err := c.fetchRepositoryPage(ctx, options)
		if err != nil {
			return nil, fmt.Errorf("fetching page %d: %w", options.Page, err)
		}

		// convert gitea repositories to repo type
		repositories := convertGiteaRepositories(giteaRepos)
		allRepositories = append(allRepositories, repositories...)

		if !hasMore {
			break
		}

		options.Page++
	}

	return allRepositories, nil
}

// fetch single page of repo
func (c *GiteaClient) fetchRepositoryPage(ctx context.Context, options GiteaAPIOptions) ([]GiteaRepository, bool, error) {
	apiURL, err := c.buildAPIURL(options)
	if err != nil {
		return nil, false, fmt.Errorf("building API URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	req.Header.Set("Accept", "application/json")

	logger.Print("Making API request to: "+apiURL, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("making request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Print("WARNING: failed to close response body: "+closeErr.Error(), nil)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	var giteaRepos []GiteaRepository
	if err := json.NewDecoder(resp.Body).Decode(&giteaRepos); err != nil {
		return nil, false, fmt.Errorf("decoding JSON response: %w", err)
	}

	// check for more pages
	hasMore := strings.Contains(resp.Header.Get("Link"), `rel="next"`)

	return giteaRepos, hasMore, nil
}

// build api url
func (c *GiteaClient) buildAPIURL(options GiteaAPIOptions) (string, error) {
	baseURL := fmt.Sprintf("https://%s/api/v1/user/repos", c.baseURL)

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parsing base URL: %w", err)
	}

	query := u.Query()
	query.Set("visibility", options.Visibility)
	query.Set("sort", options.Sort)
	query.Set("limit", strconv.Itoa(options.Limit))
	query.Set("page", strconv.Itoa(options.Page))

	// handle archived
	switch options.IncludeArchived {
	case "excluded":
		query.Set("archived", "false")
	case "only":
		query.Set("archived", "true")
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

// convert gitea repos to repo type
func convertGiteaRepositories(giteaRepos []GiteaRepository) []Repository {
	repositories := make([]Repository, len(giteaRepos))
	for i, giteaRepo := range giteaRepos {
		repositories[i] = Repository{
			Name:              giteaRepo.Name,
			PathWithNamespace: giteaRepo.FullName,
		}
	}
	return repositories
}

// connection validation
func (c *GiteaClient) ValidateConnection(ctx context.Context) error {
	apiURL := fmt.Sprintf("https://%s/api/v1/user", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("creating validation request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("making validation request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Print("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid or expired token")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API validation failed with status %d", resp.StatusCode)
	}

	return nil
}

// simply count git repos only
func (c *GiteaClient) GetRepositoryCount(ctx context.Context, options GiteaAPIOptions) (int, error) {

	repos, err := c.fetchAllRepositories(ctx, options)
	if err != nil {
		return 0, fmt.Errorf("counting repositories: %w", err)
	}

	return len(repos), nil
}
