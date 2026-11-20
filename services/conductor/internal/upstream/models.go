package upstream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	maxModels       = 1000
	maxModelNameLen = 256
	maxModelPages   = 5
)

type modelList struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
	HasMore bool   `json:"has_more"`
	LastID  string `json:"last_id"`
}

// AnthropicModels lists the models of an Anthropic API at baseURL, the URL
// without /v1.
func (c *Client) AnthropicModels(ctx context.Context, baseURL, key string) ([]string, Check) {
	var models []string

	after := ""

	for range maxModelPages {
		query := url.Values{"limit": {"1000"}}

		if after != "" {
			query.Set("after_id", after)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(baseURL, "/")+"/v1/models?"+query.Encode(), nil)

		if err != nil {
			return nil, unreachable(err)
		}

		req.Header.Set("x-api-key", key)
		req.Header.Set("anthropic-version", "2023-06-01")

		list, check := c.modelPage(req)

		if check.Status != Valid {
			return nil, check
		}

		models = appendModels(models, list)

		if !list.HasMore || list.LastID == "" || len(models) >= maxModels {
			break
		}

		after = list.LastID
	}

	return sortedModels(models), Check{Status: Valid}
}

// OpenAIModels lists the models of an OpenAI or OpenAI-compatible API at
// baseURL, the URL that ends before /models, usually in /v1.
func (c *Client) OpenAIModels(ctx context.Context, baseURL, key string) ([]string, Check) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(baseURL, "/")+"/models", nil)

	if err != nil {
		return nil, unreachable(err)
	}

	req.Header.Set("Authorization", "Bearer "+key)

	list, check := c.modelPage(req)

	if check.Status != Valid {
		return nil, check
	}

	return sortedModels(appendModels(nil, list)), Check{Status: Valid}
}

func (c *Client) modelPage(req *http.Request) (modelList, Check) {
	req.Header.Set("Accept", "application/json")

	resp, body, err := c.do(req)

	if err != nil {
		return modelList{}, unreachable(err)
	}

	if resp.StatusCode != http.StatusOK {
		return modelList{}, statusCheck(resp.StatusCode)
	}

	var list modelList

	if err := json.Unmarshal(body, &list); err != nil || list.Data == nil {
		return modelList{}, Check{Status: Unreachable, Reason: "it didn't answer with a list of models"}
	}

	return list, Check{Status: Valid}
}

func appendModels(models []string, list modelList) []string {
	for _, model := range list.Data {
		if len(models) >= maxModels {
			break
		}

		if isPrintable(model.ID, maxModelNameLen) {
			models = append(models, model.ID)
		}
	}

	return models
}

func sortedModels(models []string) []string {
	slices.Sort(models)

	return slices.Compact(models)
}

func isPrintable(s string, maxLen int) bool {
	if s == "" || len(s) > maxLen || !utf8.ValidString(s) {
		return false
	}

	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}

	return true
}
