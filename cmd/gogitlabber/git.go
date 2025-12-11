package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v6"
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
	generalErrors           []string
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
		stats.generalErrors = append(stats.generalErrors, repoPath)
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
	semaphore := make(chan struct{}, globalConfig.Concurrency)

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
	repoDestination := filepath.Join(globalConfig.Destination, repoName)

	logger.Print("Starting on repository: "+repoName, nil)

	// check if repo exists
	_, err := git.PlainOpen(repoDestination)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			// repo doesn't exist, clone it
			gitURL := buildGitURL(repoName)
			return cloneRepository(repoName, repoDestination, gitURL)
		}
		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("opening repository: %w", err),
		}
	}

	// repo exists, pull it
	return pullRepository(repoName, repoDestination)
}

// craft git url with auth embedded (for remote URL storage)
func buildGitURL(repoName string) string {
	return fmt.Sprintf("https://%s-token:%s@%s/%s.git",
		globalConfig.GitBackend, globalConfig.GitToken, globalConfig.GitHost, repoName)
}

// clone new repository
func cloneRepository(repoName, repoDestination, gitURL string) GitOperationResult {
	logger.Print("Cloning repository: "+repoName, nil)

	// ensure parent directory exists
	parentDir := filepath.Dir(repoDestination)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("creating parent directory: %w", err),
		}
	}

	_, err := git.PlainClone(repoDestination, &git.CloneOptions{
		URL:      gitURL,
		Progress: nil,
	})

	if err != nil {
		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("cloning repository: %w", err),
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

	// open repository
	repo, err := git.PlainOpen(repoDestination)
	if err != nil {
		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("opening repository: %w", err),
		}
	}

	// update remote URL with current token (in case token changed)
	gitURL := buildGitURL(repoName)
	if err := updateRemoteURL(repoDestination, gitURL); err != nil {
		logger.Print("WARNING: failed to update remote URL: "+err.Error(), nil)
	}

	// get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("getting worktree: %w", err),
		}
	}

	// check for uncommitted/unstaged changes
	status, err := worktree.Status()
	if err != nil {
		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("checking status: %w", err),
		}
	}

	if !status.IsClean() {
		// determine error type
		errorType := "unstaged"
		for _, s := range status {
			if s.Staging != git.Unmodified {
				errorType = "uncommitted"
				break
			}
		}

		return GitOperationResult{
			RepoName:  repoName,
			Operation: "error",
			Error:     fmt.Errorf("repository has local changes"),
			ErrorType: errorType,
		}
	}

	// pull changes
	err = worktree.Pull(&git.PullOptions{
		Progress: nil,
	})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			// not an error, just already up to date
			logger.Print("Repository already up to date: "+repoName, nil)
		} else {
			return GitOperationResult{
				RepoName:  repoName,
				Operation: "error",
				Error:     fmt.Errorf("pulling repository: %w", err),
				ErrorType: "other",
			}
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

// set git user config
func setGitUserConfig(repoName, repoDestination string) error {
	repo, err := git.PlainOpen(repoDestination)
	if err != nil {
		return fmt.Errorf("opening repository: %w", err)
	}

	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("getting config: %w", err)
	}

	cfg.User.Name = globalConfig.GitUserName
	cfg.User.Email = globalConfig.GitUserMail

	if err := repo.SetConfig(cfg); err != nil {
		return fmt.Errorf("setting config: %w", err)
	}

	logger.Print("Set git user config for: "+repoName, nil)
	return nil
}

// update remote URL with current token
func updateRemoteURL(repoDestination, gitURL string) error {
	repo, err := git.PlainOpen(repoDestination)
	if err != nil {
		return fmt.Errorf("opening repository: %w", err)
	}

	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("getting config: %w", err)
	}

	// update first remote's URL
	for name := range cfg.Remotes {
		cfg.Remotes[name].URLs = []string{gitURL}
		break
	}

	if err := repo.SetConfig(cfg); err != nil {
		return fmt.Errorf("setting config: %w", err)
	}

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
		switch result.ErrorType {
		case "unstaged":
			stats.IncrementCounter("unstaged", result.RepoName)
			logger.Print("Found unstaged changes in: "+result.RepoName, nil)

		case "uncommitted":
			stats.IncrementCounter("uncommitted", result.RepoName)
			logger.Print("Found uncommitted changes in: "+result.RepoName, nil)

		default:
			stats.IncrementCounter("error", result.RepoName)
			logger.Print("ERROR processing "+result.RepoName+": "+result.Error.Error(), nil)
		}
	}

	// update progress bar
	if !globalConfig.Debug {
		_ = bar.Add(1)
	}
}
