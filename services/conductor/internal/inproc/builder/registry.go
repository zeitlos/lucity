package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// registryBaseURL returns the base URL for OCI Distribution Spec API calls.
func (c *Client) registryBaseURL() string {
	scheme := "https"
	if c.registryInsecure {
		scheme = "http"
	}
	return scheme + "://" + c.registryURL
}

// registryRequest creates an HTTP request with optional Basic auth.
func (c *Client) registryRequest(ctx context.Context, method, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	if c.registryUsername != "" && c.registryPassword != "" {
		req.SetBasicAuth(c.registryUsername, c.registryPassword)
	}
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json")
	return req, nil
}

type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

type tagsResponse struct {
	Tags []string `json:"tags"`
}

// projectRepositories returns all OCI repositories belonging to a project
// by querying the catalog and filtering by the workspace/project prefix.
func (c *Client) projectRepositories(ctx context.Context, workspace, project string) ([]string, error) {
	prefix := workspace + "/" + project + "/"
	url := c.registryBaseURL() + "/v2/_catalog"

	req, err := c.registryRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create catalog request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog returned status %d", resp.StatusCode)
	}

	var catalog catalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return nil, fmt.Errorf("failed to decode catalog: %w", err)
	}

	var repos []string
	for _, repo := range catalog.Repositories {
		if strings.HasPrefix(repo, prefix) {
			repos = append(repos, repo)
		}
	}
	return repos, nil
}

// deleteRepository deletes all manifests (tags) in a repository.
func (c *Client) deleteRepository(ctx context.Context, repo string) error {
	tags, err := c.repositoryTags(ctx, repo)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		if err := c.deleteManifest(ctx, repo, tag); err != nil {
			slog.Warn("failed to delete manifest", "repo", repo, "tag", tag, "error", err)
		}
	}
	return nil
}

// repositoryTags lists all tags for a repository. Returns nil if the repo doesn't exist.
func (c *Client) repositoryTags(ctx context.Context, repo string) ([]string, error) {
	url := c.registryBaseURL() + "/v2/" + repo + "/tags/list"

	req, err := c.registryRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create tags request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tags list returned status %d", resp.StatusCode)
	}

	var tags tagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode tags: %w", err)
	}
	return tags.Tags, nil
}

// deleteManifest resolves a tag to its digest and deletes the manifest.
func (c *Client) deleteManifest(ctx context.Context, repo, tag string) error {
	// HEAD to get the digest
	headURL := c.registryBaseURL() + "/v2/" + repo + "/manifests/" + tag
	headReq, err := c.registryRequest(ctx, http.MethodHead, headURL)
	if err != nil {
		return fmt.Errorf("failed to create HEAD request: %w", err)
	}

	headResp, err := http.DefaultClient.Do(headReq)
	if err != nil {
		return fmt.Errorf("HEAD manifest failed: %w", err)
	}
	io.Copy(io.Discard, headResp.Body)
	headResp.Body.Close()

	if headResp.StatusCode == http.StatusNotFound {
		return nil // already gone
	}
	if headResp.StatusCode != http.StatusOK {
		return fmt.Errorf("HEAD manifest returned status %d", headResp.StatusCode)
	}

	digest := headResp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return fmt.Errorf("HEAD manifest returned no Docker-Content-Digest")
	}

	// DELETE by digest
	delURL := c.registryBaseURL() + "/v2/" + repo + "/manifests/" + digest
	delReq, err := c.registryRequest(ctx, http.MethodDelete, delURL)
	if err != nil {
		return fmt.Errorf("failed to create DELETE request: %w", err)
	}

	delResp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		return fmt.Errorf("DELETE manifest failed: %w", err)
	}
	io.Copy(io.Discard, delResp.Body)
	delResp.Body.Close()

	if delResp.StatusCode == http.StatusNotFound || delResp.StatusCode == http.StatusAccepted {
		return nil
	}
	if delResp.StatusCode != http.StatusOK && delResp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("DELETE manifest returned status %d", delResp.StatusCode)
	}
	return nil
}
