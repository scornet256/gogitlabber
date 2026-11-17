package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/scornet256/go-logger"
)

// version is set at build time by goreleaser (-X main.version=...)
var version = "dev"

var globalConfig *Config

// repository data
type Repository struct {
	Name              string
	PathWithNamespace string
}

func main() {

	// set appname for logger
	logger.SetAppName("gogitlabber")

	// manage all argument magic and load configuration
	globalConfig = manageArguments()

	// set debugging
	logger.SetDebug(globalConfig.Debug)

	// cancel api calls and git operations on ctrl-c / sigterm
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// make initial progressbar
	if !globalConfig.Debug {
		progressBar()
	}

	// fetch repository information
	repositories, err := fetchRepositories(ctx, globalConfig)
	if err != nil {
		fatal("fetching repositories failed", err)
	}

	if err := updateProgressBar(len(repositories)); err != nil {
		logger.Print("WARNING: failed to update progress bar", err)
	}

	// manage found repositories
	stats := &GitStats{}
	CheckoutRepositories(ctx, repositories, stats)
	printDetailedSummary(stats)

	if ctx.Err() != nil {
		fatal("interrupted", nil)
	}
}

// fetch repositories from the configured backend
func fetchRepositories(ctx context.Context, cfg *Config) ([]Repository, error) {
	var repositories []Repository
	var err error

	switch cfg.GitBackend {
	case "gitea":
		repositories, err = NewGiteaClient(cfg.GitHost, cfg.GitToken).FetchRepositories(ctx, cfg.IncludeArchived)
	case "gitlab":
		repositories, err = NewGitLabClient(cfg.GitHost, cfg.GitToken).FetchRepositories(ctx, cfg.IncludeArchived)
	default:
		return nil, fmt.Errorf("unsupported git backend: %s (supported: gitlab|gitea)", cfg.GitBackend)
	}
	if err != nil {
		return nil, err
	}

	if len(repositories) == 0 {
		return nil, errors.New("no repositories found")
	}

	logger.Print(fmt.Sprintf("Successfully fetched %d repositories", len(repositories)), nil)
	return repositories, nil
}

// print an error to stderr and exit; logger.Fatal is silent when debug is off
func fatal(message string, err error) {
	if err != nil {
		message += ": " + err.Error()
	}
	fmt.Fprintln(os.Stderr, "\ngogitlabber: "+message)
	os.Exit(1)
}
