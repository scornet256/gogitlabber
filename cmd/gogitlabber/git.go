package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/scornet256/go-logger"
)

// collect git operations results
type GitOperationResult struct {
	RepoName  string
	Operation string
	Error     error
	ErrorType string
}

// collect git stats
type GitStats struct {
	mu                      sync.Mutex
	clonedCount             int
	pulledCount             int
	errorCount              int
	pullErrorMsgUnstaged    []string
	pullErrorMsgUncommitted []string
}

// increment counters
func (stats *GitStats) IncrementCounter(operation string, repoPath string) {

	stats.mu.Lock()
	defer stats.mu.Unlock()

	switch operation {
	case "cloned":
		stats.clonedCount++
	case "pulled":
		stats.pulledCount++
	case "error":
		stats.errorCount++
	case "unstaged":
		stats.errorCount++
		stats.pullErrorMsgUnstaged = append(stats.pullErrorMsgUnstaged, repoPath)
	case "uncommitted":
		stats.errorCount++
		stats.pullErrorMsgUncommitted = append(stats.pullErrorMsgUncommitted, repoPath)
	}
}

// concurrent git operations
func CheckoutRepositories(repositories []Repository, stats *GitStats) {

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, config.Concurrency)

	for _, repo := range repositories {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(repo Repository) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			result := processRepository(repo)
			handleResult(result, stats)
		}(repo)
	}

	wg.Wait()
}

// manage single repo
func processRepository(repo Repository) GitOperationResult {
	repoName := string(repo.PathWithNamespace)
	repoDestination := filepath.Join(config.Destination, repoName)

	logger.Print("Starting on repository: "+repoName, nil)

	// check repo status
	status, err := checkRepositoryStatus(repoDestination)
	if err != nil {
		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("checking repository status: %w", err),
		}
	}

	gitURL := buildGitURL(repoName)

	switch {
	case strings.Contains(status, "No such file or directory"):
		return cloneRepository(repoName, repoDestination, gitURL)
	case strings.Contains(status, gitURL):
		return pullRepository(repoName, repoDestination)
	default:
		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("unexpected repository status: %s", status),
		}
	}
}

// check repo status
func checkRepositoryStatus(repoDestination string) (string, error) {
	cmd := exec.Command("git", "-C", repoDestination, "remote", "-v")
	output, err := cmd.CombinedOutput()

	// If directory doesn't exist, that's expected for new clones
	if err != nil && strings.Contains(string(output), "No such file or directory") {
		return string(output), nil
	}

	return string(output), err
}

// craft git url with auth
func buildGitURL(repoName string) string {
	return fmt.Sprintf("https://%s-token:%s@%s/%s.git",
		config.GitBackend, config.GitToken, config.GitHost, repoName)
}

// clone new repository
func cloneRepository(repoName, repoDestination, gitURL string) GitOperationResult {
	logger.Print("Cloning repository: "+repoName, nil)

	cmd := exec.Command("git", "clone", gitURL, repoDestination)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("cloning repository: %w, output: %s", err, string(output)),
		}
	}

	// set git user config
	if err := setGitUserConfig(repoName, repoDestination); err != nil {
		logger.Print("WARNING: failed to set git user config: "+err.Error(), nil)
	}

	return GitOperationResult{
		RepoName:  repoName,
		Operation: "cloned",
	}
}

// pull repo
func pullRepository(repoName, repoDestination string) GitOperationResult {
	logger.Print("Pulling repository: "+repoName, nil)

	// Find remote
	remote, err := findRemote(repoDestination)
	if err != nil {
		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("finding remote: %w", err),
		}
	}

	// pull changes
	cmd := exec.Command("git", "-C", repoDestination, "pull", remote)
	output, err := cmd.CombinedOutput()

	if err != nil {
		errorType := classifyPullError(string(output))
		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("pulling repository: %w, output: %s", err, string(output)),
			ErrorType: errorType,
		}
	}

	// set git user configuration
	if err := setGitUserConfig(repoName, repoDestination); err != nil {
		logger.Print("WARNING: failed to set git user config: "+err.Error(), nil)
	}

	return GitOperationResult{
		RepoName:  repoName,
		Operation: "pulled",
	}
}

// find remote for repo
func findRemote(repoDestination string) (string, error) {
	cmd := exec.Command("git", "-C", repoDestination, "remote", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("getting remote: %w", err)
	}

	remotes := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(remotes) == 0 {
		return "", fmt.Errorf("no remotes found")
	}

	return remotes[0], nil
}

// manage pull error
func classifyPullError(output string) string {
	switch {
	case strings.Contains(output, "You have unstaged changes"):
		return "unstaged"
	case strings.Contains(output, "Your index contains uncommitted changes"):
		return "uncommitted"
	default:
		return "other"
	}
}

// set git user config
func setGitUserConfig(repoName, repoDestination string) error {

	// git user name
	if err := setGitUserName(repoName, repoDestination); err != nil {
		return fmt.Errorf("setting username: %w", err)
	}

	// git user mail
	if err := setGitUserEmail(repoName, repoDestination); err != nil {
		return fmt.Errorf("setting email: %w", err)
	}

	return nil
}

// set git user name
func setGitUserName(repoName, repoDestination string) error {

	cmd := exec.Command("git", "-C", repoDestination, "config", "user.name", config.GitUserName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("setting git username: %w, output: %s", err, string(output))
	}

	logger.Print("Set git username for: "+repoName, nil)
	return nil
}

// set git user mail
func setGitUserEmail(repoName, repoDestination string) error {

	cmd := exec.Command("git", "-C", repoDestination, "config", "user.email", config.GitUserMail)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("setting git email: %w, output: %s", err, string(output))
	}

	logger.Print("Set git email for: "+repoName, nil)
	return nil
}

// manage results
func handleResult(result GitOperationResult, stats *GitStats) {

	switch result.Operation {
	case "cloned":
		stats.IncrementCounter("cloned", "")
		logger.Print("Successfully cloned: "+result.RepoName, nil)

	case "pulled":
		stats.IncrementCounter("pulled", "")
		logger.Print("Successfully pulled: "+result.RepoName, nil)

	case "error":
		if result.ErrorType == "unstaged" {
			stats.IncrementCounter("unstaged", result.RepoName)
			logger.Print("Found unstaged changes in: "+result.RepoName, nil)
		} else if result.ErrorType == "uncommitted" {
			stats.IncrementCounter("uncommitted", result.RepoName)
			logger.Print("Found uncommitted changes in: "+result.RepoName, nil)
		} else {
			stats.IncrementCounter("error", "")
			logger.Print("ERROR processing "+result.RepoName+": "+result.Error.Error(), nil)
		}
	}

	// update progress bar
	if !config.Debug {
		_ = bar.Add(1)
	}
}
