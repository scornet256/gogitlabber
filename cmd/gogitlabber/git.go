package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-git/v6"
	"github.com/scornet256/go-logger"
)

const remoteName = "origin"

// result of processing a single repository
type outcome int

const (
	outcomeCloned outcome = iota
	outcomePulled
	outcomeUnstaged
	outcomeUncommitted
	outcomeError
)

// collect git stats
type GitStats struct {
	mu          sync.Mutex
	cloned      int
	pulled      int
	unstaged    []string
	uncommitted []string
	failed      []string
}

// record the outcome of a repository
func (stats *GitStats) record(repoName string, result outcome, err error) {
	stats.mu.Lock()
	defer stats.mu.Unlock()

	switch result {
	case outcomeCloned:
		stats.cloned++
	case outcomePulled:
		stats.pulled++
	case outcomeUnstaged:
		stats.unstaged = append(stats.unstaged, repoName)
	case outcomeUncommitted:
		stats.uncommitted = append(stats.uncommitted, repoName)
	case outcomeError:
		stats.failed = append(stats.failed, fmt.Sprintf("%s: %v", repoName, err))
	}
}

func (stats *GitStats) errorCount() int {
	return len(stats.unstaged) + len(stats.uncommitted) + len(stats.failed)
}

// concurrent git operations
func CheckoutRepositories(ctx context.Context, repositories []Repository, stats *GitStats) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, globalConfig.Concurrency)

loop:
	for _, repo := range repositories {
		select {
		case semaphore <- struct{}{}:
		case <-ctx.Done():
			break loop
		}

		wg.Go(func() {
			defer func() { <-semaphore }()

			result, err := processRepository(ctx, repo)
			stats.record(repo.PathWithNamespace, result, err)
			logResult(repo.PathWithNamespace, result, err)

			if bar != nil {
				_ = bar.Add(1)
			}
		})
	}

	wg.Wait()
}

// clone or pull a single repository
func processRepository(ctx context.Context, repo Repository) (outcome, error) {
	repoName := repo.PathWithNamespace

	// refuse names that would escape the destination directory
	if !filepath.IsLocal(filepath.FromSlash(repoName)) {
		return outcomeError, fmt.Errorf("invalid repository path %q", repoName)
	}
	repoDestination := filepath.Join(globalConfig.Destination, filepath.FromSlash(repoName))

	logger.Print("Starting on repository: "+repoName, nil)

	gitRepo, err := git.PlainOpen(repoDestination)
	switch {
	case errors.Is(err, git.ErrRepositoryNotExists):
		return cloneRepository(ctx, repoName, repoDestination)
	case err != nil:
		return outcomeError, fmt.Errorf("opening repository: %w", err)
	}

	return pullRepository(ctx, gitRepo, repoName, repoDestination)
}

// craft the remote url: ssh without credentials, or https with the token embedded (so plain git works too)
func buildGitURL(repoName string) string {
	if useSSH() {
		return fmt.Sprintf("ssh://%s@%s:%d/%s.git",
			globalConfig.GitSSHUser, globalConfig.GitHost, globalConfig.GitSSHPort, repoName)
	}
	return fmt.Sprintf("https://%s-token:%s@%s/%s.git",
		globalConfig.GitBackend, globalConfig.GitToken, globalConfig.GitHost, repoName)
}

// ssh goes through the git cli, so openssh applies the full ~/.ssh/config (proxyjump, identityfile, known hosts)
func useSSH() bool {
	return globalConfig.GitProtocol == "ssh"
}

// run the git cli without prompts; stderr ends up in the error
func runGit(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	// fail instead of asking for a passphrase or host key behind the progress bar
	if os.Getenv("GIT_SSH_COMMAND") == "" {
		cmd.Env = append(cmd.Env, "GIT_SSH_COMMAND=ssh -o BatchMode=yes")
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("%w: %s", err, msg)
		}
		return err
	}
	return nil
}

// fast-forward pull with the git cli (--ff-only like go-git, and it overrides a global pull.rebase)
func pullWithGit(ctx context.Context, repoDestination string) error {
	args := []string{"-C", repoDestination, "pull", "--quiet", "--ff-only", remoteName}

	// without an upstream, pull the remote HEAD like go-git does
	if runGit(ctx, "-C", repoDestination, "rev-parse", "--verify", "--quiet", "@{upstream}") != nil {
		args = append(args, "HEAD")
	}
	return runGit(ctx, args...)
}

// clone new repository
func cloneRepository(ctx context.Context, repoName, repoDestination string) (outcome, error) {
	logger.Print("Cloning repository: "+repoName, nil)

	if err := os.MkdirAll(filepath.Dir(repoDestination), 0o755); err != nil {
		return outcomeError, fmt.Errorf("creating parent directory: %w", err)
	}

	var gitRepo *git.Repository
	var err error
	if useSSH() {
		err = runGit(ctx, "clone", "--quiet", "--origin", remoteName, "--", buildGitURL(repoName), repoDestination)
		if err == nil {
			gitRepo, err = git.PlainOpen(repoDestination)
		}
	} else {
		gitRepo, err = git.PlainCloneContext(ctx, repoDestination, &git.CloneOptions{
			URL:        buildGitURL(repoName),
			RemoteName: remoteName,
		})
	}
	if err != nil {
		return outcomeError, fmt.Errorf("cloning repository: %w", err)
	}

	if err := setGitUserConfig(gitRepo); err != nil {
		logger.Print("WARNING: failed to set git user config for "+repoName, err)
	}

	return outcomeCloned, nil
}

// pull existing repository
func pullRepository(ctx context.Context, gitRepo *git.Repository, repoName, repoDestination string) (outcome, error) {
	logger.Print("Pulling repository: "+repoName, nil)

	// update remote url with current token (in case token changed)
	if err := setRemoteURL(gitRepo, buildGitURL(repoName)); err != nil {
		logger.Print("WARNING: failed to update remote URL for "+repoName, err)
	}

	worktree, err := gitRepo.Worktree()
	if err != nil {
		return outcomeError, fmt.Errorf("getting worktree: %w", err)
	}

	// do not touch repositories with local changes
	status, err := worktree.Status()
	if err != nil {
		return outcomeError, fmt.Errorf("checking status: %w", err)
	}
	if !status.IsClean() {
		return classifyStatus(status), nil
	}

	if useSSH() {
		err = pullWithGit(ctx, repoDestination)
	} else {
		err = worktree.PullContext(ctx, &git.PullOptions{RemoteName: remoteName})
	}
	switch {
	case errors.Is(err, git.NoErrAlreadyUpToDate):
		logger.Print("Repository already up to date: "+repoName, nil)
	case err != nil:
		return outcomeError, fmt.Errorf("pulling repository: %w", err)
	}

	if err := setGitUserConfig(gitRepo); err != nil {
		logger.Print("WARNING: failed to set git user config for "+repoName, err)
	}

	return outcomePulled, nil
}

// staged changes are uncommitted; worktree-only and untracked changes are unstaged
func classifyStatus(status git.Status) outcome {
	for _, s := range status {
		if s.Staging != git.Unmodified && s.Staging != git.Untracked {
			return outcomeUncommitted
		}
	}
	return outcomeUnstaged
}

// set git user config, leaving fields that are not configured alone
func setGitUserConfig(gitRepo *git.Repository) error {
	if globalConfig.GitUserName == "" && globalConfig.GitUserMail == "" {
		return nil
	}

	cfg, err := gitRepo.Config()
	if err != nil {
		return fmt.Errorf("getting config: %w", err)
	}

	if globalConfig.GitUserName != "" {
		cfg.User.Name = globalConfig.GitUserName
	}
	if globalConfig.GitUserMail != "" {
		cfg.User.Email = globalConfig.GitUserMail
	}

	return gitRepo.SetConfig(cfg)
}

// update the origin remote url; other remotes are left alone
func setRemoteURL(gitRepo *git.Repository, gitURL string) error {
	cfg, err := gitRepo.Config()
	if err != nil {
		return fmt.Errorf("getting config: %w", err)
	}

	remote, ok := cfg.Remotes[remoteName]
	if !ok {
		return fmt.Errorf("remote %q not found", remoteName)
	}
	if len(remote.URLs) == 1 && remote.URLs[0] == gitURL {
		return nil
	}
	remote.URLs = []string{gitURL}

	return gitRepo.SetConfig(cfg)
}

// debug log the outcome of a repository
func logResult(repoName string, result outcome, err error) {
	switch result {
	case outcomeCloned:
		logger.Print("Successfully cloned: "+repoName, nil)
	case outcomePulled:
		logger.Print("Successfully pulled: "+repoName, nil)
	case outcomeUnstaged:
		logger.Print("Found unstaged changes in: "+repoName, nil)
	case outcomeUncommitted:
		logger.Print("Found uncommitted changes in: "+repoName, nil)
	case outcomeError:
		logger.Print("ERROR processing "+repoName, err)
	}
}
