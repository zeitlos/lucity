package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ImageSearchResult represents a container image from Docker Hub.
type ImageSearchResult struct {
	Name        string
	Description string
	StarCount   int
	PullCount   int64
	Official    bool
}

// dockerHubResponse is the raw response from Docker Hub's search API.
type dockerHubResponse struct {
	Results []struct {
		RepoName         string `json:"repo_name"`
		ShortDescription string `json:"short_description"`
		StarCount        int    `json:"star_count"`
		PullCount        int64  `json:"pull_count"`
		IsOfficial       bool   `json:"is_official"`
	} `json:"results"`
}

// SearchImages queries Docker Hub's public search API for container images.
func (c *Client) SearchImages(ctx context.Context, query string) ([]ImageSearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	u := fmt.Sprintf("https://hub.docker.com/v2/search/repositories/?query=%s&page_size=10", url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search Docker Hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Docker Hub returned status %d", resp.StatusCode)
	}

	var body dockerHubResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode Docker Hub response: %w", err)
	}

	results := make([]ImageSearchResult, 0, len(body.Results))
	for _, r := range body.Results {
		results = append(results, ImageSearchResult{
			Name:        r.RepoName,
			Description: r.ShortDescription,
			StarCount:   r.StarCount,
			PullCount:   r.PullCount,
			Official:    r.IsOfficial,
		})
	}
	return results, nil
}
