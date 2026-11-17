package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/scornet256/go-logger"
)

const apiTimeout = 30 * time.Second

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: apiTimeout}
}

// getJSON performs a GET request and decodes the JSON body into out
func getJSON(ctx context.Context, client *http.Client, apiURL string, headers map[string]string, out any) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	logger.Print("Making API request to: "+apiURL, nil)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("API request failed with status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return nil, fmt.Errorf("decoding JSON response: %w", err)
	}

	return resp.Header, nil
}

// archived query value for the include_archived option; empty means no filter
func archivedParam(includeArchived string) string {
	switch includeArchived {
	case "excluded":
		return "false"
	case "exclusive":
		return "true"
	}
	return ""
}

// client side archived filter, in case the api ignores the query parameter
func keepArchived(archived bool, includeArchived string) bool {
	switch includeArchived {
	case "excluded":
		return !archived
	case "exclusive":
		return archived
	}
	return true
}
