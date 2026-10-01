package bunny

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const endpoint = "https://api.bunny.net"

type Client struct {
	apiKey string
	http   *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

type APIError struct {
	Status  int
	Key     string `json:"ErrorKey"`
	Message string `json:"Message"`
	Body    string
}

func (e *APIError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("bunny api error %d: %s: %s", e.Status, e.Key, e.Message)
	}

	return fmt.Sprintf("bunny api error %d: %s", e.Status, e.Body)
}

func IsNotFound(err error) bool {
	var apiErr *APIError

	if !errors.As(err, &apiErr) {
		return false
	}

	return apiErr.Status == http.StatusNotFound || strings.HasSuffix(strings.ToLower(apiErr.Key), "not_found")
}

func IsAlreadyExists(err error) bool {
	var apiErr *APIError

	if !errors.As(err, &apiErr) {
		return false
	}

	key := strings.ToLower(apiErr.Key)

	return strings.HasSuffix(key, "already_registered") || strings.HasSuffix(key, "name_taken")
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader

	if body != nil {
		data, err := json.Marshal(body)

		if err != nil {
			return err
		}

		reader = bytes.NewReader(data)
	}

	request, err := http.NewRequestWithContext(ctx, method, endpoint+path, reader)

	if err != nil {
		return err
	}

	request.Header.Set("AccessKey", c.apiKey)
	request.Header.Set("Accept", "application/json")

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.http.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	data, _ := io.ReadAll(response.Body)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		apiErr := &APIError{Status: response.StatusCode, Body: string(data)}
		_ = json.Unmarshal(data, apiErr)

		return apiErr
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
