package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v6"
	gitconfig "github.com/go-git/go-git/v6/config"
)

func initRepo(t *testing.T) (*git.Repository, string) {
	t.Helper()
	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	return repo, dir
}

func withConfig(t *testing.T, cfg *Config) {
	t.Helper()
	old := globalConfig
	globalConfig = cfg
	t.Cleanup(func() { globalConfig = old })
}

func TestSetRemoteURLOnlyTouchesOrigin(t *testing.T) {
	repo, _ := initRepo(t)
	for name, u := range map[string]string{"origin": "https://old/origin.git", "upstream": "https://upstream.git"} {
		if _, err := repo.CreateRemote(&gitconfig.RemoteConfig{Name: name, URLs: []string{u}}); err != nil {
			t.Fatal(err)
		}
	}

	if err := setRemoteURL(repo, "https://new/origin.git"); err != nil {
		t.Fatal(err)
	}

	cfg, err := repo.Config()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Remotes["origin"].URLs[0]; got != "https://new/origin.git" {
		t.Errorf("origin url = %s", got)
	}
	if got := cfg.Remotes["upstream"].URLs[0]; got != "https://upstream.git" {
		t.Errorf("upstream url changed to %s", got)
	}
}

func TestSetRemoteURLMissingOrigin(t *testing.T) {
	repo, _ := initRepo(t)
	if err := setRemoteURL(repo, "https://x.git"); err == nil {
		t.Fatal("expected error when origin is missing")
	}
}

func TestSetGitUserConfig(t *testing.T) {
	repo, _ := initRepo(t)

	withConfig(t, &Config{GitUserName: "Jane"})
	if err := setGitUserConfig(repo); err != nil {
		t.Fatal(err)
	}
	withConfig(t, &Config{GitUserMail: "jane@example.com"})
	if err := setGitUserConfig(repo); err != nil {
		t.Fatal(err)
	}

	cfg, err := repo.Config()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.User.Name != "Jane" || cfg.User.Email != "jane@example.com" {
		t.Errorf("user config = %q <%q>", cfg.User.Name, cfg.User.Email)
	}
}

func TestClassifyStatus(t *testing.T) {
	repo, dir := initRepo(t)
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	status := func() git.Status {
		t.Helper()
		s, err := worktree.Status()
		if err != nil {
			t.Fatal(err)
		}
		return s
	}

	// untracked file only
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := classifyStatus(status()); got != outcomeUnstaged {
		t.Errorf("untracked file: got %v, want unstaged", got)
	}

	// staged file
	if _, err := worktree.Add("file.txt"); err != nil {
		t.Fatal(err)
	}
	if got := classifyStatus(status()); got != outcomeUncommitted {
		t.Errorf("staged file: got %v, want uncommitted", got)
	}
}

func TestProcessRepositoryRejectsEscapingPath(t *testing.T) {
	withConfig(t, &Config{Destination: t.TempDir()})

	result, err := processRepository(context.Background(), Repository{PathWithNamespace: "../escape"})
	if result != outcomeError || err == nil {
		t.Fatalf("got %v, %v; want error", result, err)
	}
}

func TestGitStatsRecord(t *testing.T) {
	stats := &GitStats{}
	stats.record("a", outcomeCloned, nil)
	stats.record("b", outcomePulled, nil)
	stats.record("c", outcomeUnstaged, nil)
	stats.record("d", outcomeUncommitted, nil)
	stats.record("e", outcomeError, errors.New("boom"))

	if stats.cloned != 1 || stats.pulled != 1 || stats.errorCount() != 3 {
		t.Errorf("unexpected stats: %+v", stats)
	}
	if stats.failed[0] != "e: boom" {
		t.Errorf("failed entry = %q", stats.failed[0])
	}
}
