package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestBuildGitURL(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want string
	}{
		{
			"https",
			&Config{GitBackend: "gitea", GitHost: "git.example.com", GitProtocol: "https", GitToken: "secret"},
			"https://gitea-token:secret@git.example.com/org/repo.git",
		},
		{
			"ssh",
			&Config{GitBackend: "gitea", GitHost: "git.example.com", GitProtocol: "ssh", GitSSHPort: 2222,
				GitSSHUser: "git", GitToken: "secret"},
			"ssh://git@git.example.com:2222/org/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withConfig(t, tt.cfg)
			if got := buildGitURL("org/repo"); got != tt.want {
				t.Errorf("buildGitURL = %q, want %q", got, tt.want)
			}
		})
	}
}

// a bare remote with one commit on main, plus a clone of it
func gitFixture(t *testing.T) (remote, clone string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git cli not installed")
	}

	ctx := context.Background()
	remote = filepath.Join(t.TempDir(), "remote.git")
	work := filepath.Join(t.TempDir(), "work")
	clone = filepath.Join(t.TempDir(), "clone")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_AUTHOR_NAME", "t")
	t.Setenv("GIT_AUTHOR_EMAIL", "t@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "t")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@example.com")

	for _, args := range [][]string{
		{"init", "--quiet", "--bare", "--initial-branch=main", remote},
		{"init", "--quiet", "--initial-branch=main", work},
		{"-C", work, "commit", "--quiet", "--allow-empty", "-m", "one"},
		{"-C", work, "push", "--quiet", remote, "main"},
		{"clone", "--quiet", remote, clone},
		{"-C", work, "commit", "--quiet", "--allow-empty", "-m", "two"},
		{"-C", work, "push", "--quiet", remote, "main"},
	} {
		if err := runGit(ctx, args...); err != nil {
			t.Fatal(err)
		}
	}
	return remote, clone
}

func headOf(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func TestPullWithGit(t *testing.T) {
	remote, clone := gitFixture(t)

	if err := pullWithGit(context.Background(), clone); err != nil {
		t.Fatal(err)
	}
	if got, want := headOf(t, clone), headOf(t, remote); got != want {
		t.Errorf("head = %s, want %s", got, want)
	}
}

func TestPullWithGitWithoutUpstream(t *testing.T) {
	remote, clone := gitFixture(t)
	if err := runGit(context.Background(), "-C", clone, "branch", "--unset-upstream"); err != nil {
		t.Fatal(err)
	}

	if err := pullWithGit(context.Background(), clone); err != nil {
		t.Fatal(err)
	}
	if got, want := headOf(t, clone), headOf(t, remote); got != want {
		t.Errorf("head = %s, want %s", got, want)
	}
}

func TestRunGitReportsStderr(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git cli not installed")
	}

	err := runGit(context.Background(), "-C", t.TempDir(), "pull", "--ff-only", "origin")
	if err == nil {
		t.Fatal("expected error outside a repository")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("stderr missing from error: %v", err)
	}
}
