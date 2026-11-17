package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validConfig() *Config {
	cfg := defaultConfig()
	cfg.GitBackend = "gitlab"
	cfg.GitToken = "token"
	return cfg
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr string
	}{
		{"valid", func(*Config) {}, ""},
		{"gitea", func(c *Config) { c.GitBackend = "gitea" }, ""},
		{"missing backend", func(c *Config) { c.GitBackend = "" }, "git_backend is required"},
		{"bad backend", func(c *Config) { c.GitBackend = "github" }, "unsupported git_backend"},
		{"missing host", func(c *Config) { c.GitHost = "" }, "git_host is required"},
		{"missing token", func(c *Config) { c.GitToken = "" }, "git_token is required"},
		{"archived any", func(c *Config) { c.IncludeArchived = "any" }, ""},
		{"archived exclusive", func(c *Config) { c.IncludeArchived = "exclusive" }, ""},
		{"archived bad", func(c *Config) { c.IncludeArchived = "only" }, "invalid include_archived"},
		{"zero concurrency", func(c *Config) { c.Concurrency = 0 }, "concurrency"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.modify(cfg)
			err := cfg.validateConfig()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got error %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GGL_TEST", "sub")

	tests := map[string]string{
		"~":                      home,
		"~/code":                 filepath.Join(home, "code"),
		"$HOME/Documents":        filepath.Join(home, "Documents"),
		"/tmp/$GGL_TEST/":        "/tmp/sub",
		"relative/../dir":        "dir",
		"~other/not-expanded":    "~other/not-expanded",
		"~/.config/gogitlabber/": filepath.Join(home, ".config", "gogitlabber"),
	}

	for in, want := range tests {
		if got := expandPath(in); got != want {
			t.Errorf("expandPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestProcessConfigNormalizesHost(t *testing.T) {
	for _, host := range []string{"gitlab.example.com", "https://gitlab.example.com/", "http://gitlab.example.com"} {
		cfg := validConfig()
		cfg.GitHost = host
		cfg.processConfig()
		if cfg.GitHost != "gitlab.example.com" {
			t.Errorf("host %q normalized to %q", host, cfg.GitHost)
		}
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfig(t *testing.T) {
	cfg, err := loadConfig(writeConfig(t, "git_backend: gitea\ngit_token: abc\nconcurrency: 3\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GitBackend != "gitea" || cfg.GitToken != "abc" || cfg.Concurrency != 3 {
		t.Errorf("unexpected config: %+v", cfg)
	}
	// defaults are kept for missing keys
	if cfg.GitHost != "gitlab.com" || cfg.IncludeArchived != "excluded" {
		t.Errorf("defaults not applied: %+v", cfg)
	}
}

func TestLoadConfigEmptyFile(t *testing.T) {
	if _, err := loadConfig(writeConfig(t, "")); err != nil {
		t.Fatalf("empty file should load defaults, got %v", err)
	}
}

func TestLoadConfigUnknownKey(t *testing.T) {
	_, err := loadConfig(writeConfig(t, "include_archive: any\n"))
	if err == nil || !strings.Contains(err.Error(), "include_archive") {
		t.Fatalf("expected unknown key error, got %v", err)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	if _, err := loadConfig(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("expected error for missing file")
	}
}
