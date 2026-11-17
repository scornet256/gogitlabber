package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func repoNames(repos []Repository) []string {
	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = r.PathWithNamespace
	}
	return names
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Error(err)
	}
}

// gitlab server with two pages and no X-Total-Pages header (as for >10k results)
func newGitLabServer(t *testing.T) *httptest.Server {
	pages := map[int][]gitLabProject{
		1: {{Name: "a", PathWithNamespace: "group/a"}, {Name: "old", PathWithNamespace: "group/old", Archived: true}},
		2: {{Name: "b", PathWithNamespace: "group/sub/b"}},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("PRIVATE-TOKEN") != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"401 Unauthorized"}`))
			return
		}
		if r.URL.Path != "/api/v4/projects" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < len(pages) {
			w.Header().Set("X-Next-Page", strconv.Itoa(page+1))
		} else {
			w.Header().Set("X-Next-Page", "")
		}
		writeJSON(t, w, pages[page])
	}))
}

func TestGitLabFetchRepositories(t *testing.T) {
	srv := newGitLabServer(t)
	defer srv.Close()

	tests := map[string][]string{
		"any":       {"group/a", "group/old", "group/sub/b"},
		"excluded":  {"group/a", "group/sub/b"},
		"exclusive": {"group/old"},
	}

	for mode, want := range tests {
		client := &GitLabClient{httpClient: srv.Client(), apiBase: srv.URL + "/api/v4", token: "secret"}
		repos, err := client.FetchRepositories(context.Background(), mode)
		if err != nil {
			t.Fatalf("%s: %v", mode, err)
		}
		if got := repoNames(repos); !slices.Equal(got, want) {
			t.Errorf("%s: got %v, want %v", mode, got, want)
		}
	}
}

func TestGitLabFetchRepositoriesUnauthorized(t *testing.T) {
	srv := newGitLabServer(t)
	defer srv.Close()

	client := &GitLabClient{httpClient: srv.Client(), apiBase: srv.URL + "/api/v4", token: "wrong"}
	_, err := client.FetchRepositories(context.Background(), "any")
	if err == nil || !strings.Contains(err.Error(), "401 Unauthorized") {
		t.Fatalf("expected 401 error with body, got %v", err)
	}
}

func TestGitLabProjectsURL(t *testing.T) {
	client := NewGitLabClient("gitlab.example.com", "x")

	got := client.projectsURL(3, "exclusive")
	for _, want := range []string{"https://gitlab.example.com/api/v4/projects?", "page=3", "archived=true", "membership=true"} {
		if !strings.Contains(got, want) {
			t.Errorf("url %s missing %s", got, want)
		}
	}

	if got := client.projectsURL(1, "any"); strings.Contains(got, "archived=") {
		t.Errorf("url %s should not filter archived", got)
	}
}

func TestGiteaFetchRepositories(t *testing.T) {
	pages := map[int][]giteaRepository{
		1: {{Name: "a", FullName: "org/a"}, {Name: "old", FullName: "org/old", Archived: true}},
		2: {{Name: "b", FullName: "user/b"}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "token secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/api/v1/user/repos" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < len(pages) {
			w.Header().Set("Link", fmt.Sprintf(`<%s/api/v1/user/repos?page=%d>; rel="next"`, "http://"+r.Host, page+1))
		}
		writeJSON(t, w, pages[page])
	}))
	defer srv.Close()

	tests := map[string][]string{
		"any":       {"org/a", "org/old", "user/b"},
		"excluded":  {"org/a", "user/b"},
		"exclusive": {"org/old"},
	}

	for mode, want := range tests {
		client := &GiteaClient{httpClient: srv.Client(), apiBase: srv.URL + "/api/v1", token: "secret"}
		repos, err := client.FetchRepositories(context.Background(), mode)
		if err != nil {
			t.Fatalf("%s: %v", mode, err)
		}
		if got := repoNames(repos); !slices.Equal(got, want) {
			t.Errorf("%s: got %v, want %v", mode, got, want)
		}
	}
}

func TestFetchRepositoriesUnsupportedBackend(t *testing.T) {
	cfg := validConfig()
	cfg.GitBackend = "github"
	if _, err := fetchRepositories(context.Background(), cfg); err == nil {
		t.Fatal("expected error for unsupported backend")
	}
}
