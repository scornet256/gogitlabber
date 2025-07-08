package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/scornet256/go-logger"
)

// GitLabClient encapsulates the GitLab API client functionality
type GitLabClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// GitLabProject represents a project from GitLab API
type GitLabProject struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	Archived          bool   `json:"archived"`
	LastActivityAt    string `json:"last_activity_at"`
	WebURL            string `json:"web_url"`
}

// GitLabAPIOptions holds the API request parameters
type GitLabAPIOptions struct {
	Membership      bool
	IncludeArchived string
	OrderBy         string
	Sort            string
	PerPage         int
	Page            int
	MinAccessLevel  int // 10=Guest, 20=Reporter, 30=Developer, 40=Maintainer, 50=Owner
}

// gitlab pagination info
type GitLabPaginationInfo struct {
	TotalPages   int
	TotalItems   int
	CurrentPage  int
	NextPage     int
	PreviousPage int
}

// gitlab client
func NewGitLabClient(baseURL, token string) *GitLabClient {
	return &GitLabClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
		token:   token,
	}
}

// fetch gitlab repos
func FetchRepositoriesGitLab() ([]Repository, error) {
	client := NewGitLabClient(config.GitHost, config.GitToken)

	options := GitLabAPIOptions{
		Membership:      true,
		IncludeArchived: config.IncludeArchived,
		OrderBy:         "name",
		Sort:            "asc",
		PerPage:         100,
		Page:            1,
		MinAccessLevel:  20,
	}

	repositories, err := client.fetchAllProjects(context.Background(), options)
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
func (c *GitLabClient) fetchAllProjects(ctx context.Context, options GitLabAPIOptions) ([]Repository, error) {
	var allRepositories []Repository

	for {
		gitlabProjects, pagination, err := c.fetchProjectPage(ctx, options)
		if err != nil {
			return nil, fmt.Errorf("fetching page %d: %w", options.Page, err)
		}

		// convert gitlab repositories to repo type
		repositories := convertGitLabProjects(gitlabProjects, options.IncludeArchived)
		allRepositories = append(allRepositories, repositories...)

		logger.Print(fmt.Sprintf("Fetched page %d/%d (%d projects)",
			pagination.CurrentPage, pagination.TotalPages, len(gitlabProjects)), nil)

		// check if we have more pages
		if pagination.NextPage == 0 || pagination.CurrentPage >= pagination.TotalPages {
			break
		}

		options.Page = pagination.NextPage
	}

	return allRepositories, nil
}

// fetch single page of repo
func (c *GitLabClient) fetchProjectPage(ctx context.Context, options GitLabAPIOptions) ([]GitLabProject, GitLabPaginationInfo, error) {
	apiURL, err := c.buildAPIURL(options)
	if err != nil {
		return nil, GitLabPaginationInfo{}, fmt.Errorf("building API URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, GitLabPaginationInfo{}, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Accept", "application/json")

	logger.Print("Making API request to: "+apiURL, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, GitLabPaginationInfo{}, fmt.Errorf("making request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Print("WARNING: failed to close response body: "+closeErr.Error(), nil)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, GitLabPaginationInfo{}, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	var gitlabProjects []GitLabProject
	if err := json.NewDecoder(resp.Body).Decode(&gitlabProjects); err != nil {
		return nil, GitLabPaginationInfo{}, fmt.Errorf("decoding JSON response: %w", err)
	}

	// check for more pages
	pagination := parsePaginationHeaders(resp.Header)

	return gitlabProjects, pagination, nil
}

// build final api url
func (c *GitLabClient) buildAPIURL(options GitLabAPIOptions) (string, error) {
	baseURL := fmt.Sprintf("https://%s/api/v4/projects", c.baseURL)

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parsing base URL: %w", err)
	}

	query := u.Query()

	if options.Membership {
		query.Set("membership", "true")
	}

	query.Set("order_by", options.OrderBy)
	query.Set("sort", options.Sort)
	query.Set("per_page", strconv.Itoa(options.PerPage))
	query.Set("page", strconv.Itoa(options.Page))

	if options.MinAccessLevel > 0 {
		query.Set("min_access_level", strconv.Itoa(options.MinAccessLevel))
	}

	// handle archived
	switch options.IncludeArchived {
	case "excluded":
		query.Set("archived", "false")
	case "only":
		query.Set("archived", "true")
		// For "included" or any other value, don't set the archived parameter
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

// parse pagination headers
func parsePaginationHeaders(headers http.Header) GitLabPaginationInfo {

	pagination := GitLabPaginationInfo{}

	if totalPages := headers.Get("X-Total-Pages"); totalPages != "" {
		pagination.TotalPages, _ = strconv.Atoi(totalPages)
	}

	if totalItems := headers.Get("X-Total"); totalItems != "" {
		pagination.TotalItems, _ = strconv.Atoi(totalItems)
	}

	if currentPage := headers.Get("X-Page"); currentPage != "" {
		pagination.CurrentPage, _ = strconv.Atoi(currentPage)
	}

	if nextPage := headers.Get("X-Next-Page"); nextPage != "" {
		pagination.NextPage, _ = strconv.Atoi(nextPage)
	}

	if prevPage := headers.Get("X-Prev-Page"); prevPage != "" {
		pagination.PreviousPage, _ = strconv.Atoi(prevPage)
	}

	return pagination
}

// convert gitlab repos to repo type
func convertGitLabProjects(gitlabProjects []GitLabProject, includeArchived string) []Repository {
	var repositories []Repository

	for _, project := range gitlabProjects {
		// Additional filtering based on archived status if needed
		if includeArchived == "excluded" && project.Archived {
			continue
		}
		if includeArchived == "only" && !project.Archived {
			continue
		}

		repositories = append(repositories, Repository{
			Name:              project.Name,
			PathWithNamespace: project.PathWithNamespace,
		})
	}

	return repositories
}

// connection validation
func (c *GitLabClient) ValidateConnection(ctx context.Context) error {
	apiURL := fmt.Sprintf("https://%s/api/v4/user", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("creating validation request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
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

// fetch projects group
func (c *GitLabClient) GetProjectsByGroup(ctx context.Context, groupID string, options GitLabAPIOptions) ([]Repository, error) {
	var allRepositories []Repository

	for {
		gitlabProjects, pagination, err := c.fetchGroupProjectPage(ctx, groupID, options)
		if err != nil {
			return nil, fmt.Errorf("fetching group page %d: %w", options.Page, err)
		}

		// Convert GitLab projects to our Repository type
		repositories := convertGitLabProjects(gitlabProjects, options.IncludeArchived)
		allRepositories = append(allRepositories, repositories...)

		logger.Print(fmt.Sprintf("Fetched group page %d/%d (%d projects)",
			pagination.CurrentPage, pagination.TotalPages, len(gitlabProjects)), nil)

		// Check if we have more pages
		if pagination.NextPage == 0 || pagination.CurrentPage >= pagination.TotalPages {
			break
		}

		options.Page = pagination.NextPage
	}

	return allRepositories, nil
}

// fetch project page
func (c *GitLabClient) fetchGroupProjectPage(ctx context.Context, groupID string, options GitLabAPIOptions) ([]GitLabProject, GitLabPaginationInfo, error) {
	apiURL, err := c.buildGroupAPIURL(groupID, options)
	if err != nil {
		return nil, GitLabPaginationInfo{}, fmt.Errorf("building group API URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, GitLabPaginationInfo{}, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Accept", "application/json")

	logger.Print("Making group API request to: "+apiURL, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, GitLabPaginationInfo{}, fmt.Errorf("making request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Print("WARNING: failed to close response body: "+closeErr.Error(), nil)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, GitLabPaginationInfo{}, fmt.Errorf("group API request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	var gitlabProjects []GitLabProject
	if err := json.NewDecoder(resp.Body).Decode(&gitlabProjects); err != nil {
		return nil, GitLabPaginationInfo{}, fmt.Errorf("decoding JSON response: %w", err)
	}

	pagination := parsePaginationHeaders(resp.Header)

	return gitlabProjects, pagination, nil
}

// build api url
func (c *GitLabClient) buildGroupAPIURL(groupID string, options GitLabAPIOptions) (string, error) {
	baseURL := fmt.Sprintf("https://%s/api/v4/groups/%s/projects", c.baseURL, groupID)

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parsing base URL: %w", err)
	}

	query := u.Query()
	query.Set("per_page", strconv.Itoa(options.PerPage))
	query.Set("page", strconv.Itoa(options.Page))

	if options.MinAccessLevel > 0 {
		query.Set("min_access_level", strconv.Itoa(options.MinAccessLevel))
	}

	// Handle archived parameter
	switch options.IncludeArchived {
	case "excluded":
		query.Set("archived", "false")
	case "only":
		query.Set("archived", "true")
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

// get project stats
func (c *GitLabClient) GetProjectStatistics(ctx context.Context) (map[string]int, error) {
	options := GitLabAPIOptions{
		Membership: true,
		PerPage:    100,
		Page:       1,
	}

	projects, err := c.fetchAllProjects(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("fetching projects for statistics: %w", err)
	}

	stats := map[string]int{
		"total":    len(projects),
		"archived": 0,
		"active":   0,
	}

	stats["active"] = len(projects)

	return stats, nil
}
