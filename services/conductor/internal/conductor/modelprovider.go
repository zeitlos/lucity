package conductor

import (
	"context"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/upstream"
)

const (
	anthropicURL = "https://api.anthropic.com"
	openAIURL    = "https://api.openai.com/v1"
)

type ModelProvider struct {
	ID        ModelProviderID
	API       values.ModelAPI
	URL       string
	KeyHint   string
	Status    values.CredentialStatus
	CheckedAt *time.Time
	Models    []string
}

func (c *Client) workspaceRelease(ctx context.Context, workspace string) (*values.Workspace, error) {
	if !c.Capabilities(ctx, workspace).Agents {
		return values.NewWorkspace(workspace), nil
	}

	return c.deployer.Workspaces().Workspace(ctx, workspace)
}

func (c *Client) callerWorkspaceRelease(ctx context.Context) (string, *values.Workspace, error) {
	workspace, err := tenant.FromContext(ctx)

	if err != nil {
		return "", nil, err
	}

	ws, err := c.workspaceRelease(ctx, workspace)

	return workspace, ws, err
}

func (c *Client) ModelProviders(ctx context.Context) ([]ModelProvider, error) {
	workspace, ws, err := c.callerWorkspaceRelease(ctx)

	if err != nil {
		return nil, err
	}

	return modelProviders(workspace, ws), nil
}

func modelProviders(workspace string, ws *values.Workspace) []ModelProvider {
	providers := make([]ModelProvider, 0, len(ws.ModelProviders))

	for name, provider := range ws.ModelProviders {
		providers = append(providers, toModelProvider(ModelProviderID{Workspace: workspace, Name: name}, provider))
	}

	slices.SortFunc(providers, func(a, b ModelProvider) int { return strings.Compare(a.ID.Name, b.ID.Name) })

	return providers
}

func (c *Client) ModelProvider(ctx context.Context, id ModelProviderID) (*ModelProvider, error) {
	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	provider, ok := ws.ModelProviders[id.Name]

	if !ok {
		return nil, usererror.Errorf("model provider %q not found", id.Name)
	}

	return new(toModelProvider(id, provider)), nil
}

func toModelProvider(id ModelProviderID, provider values.ModelProvider) ModelProvider {
	return ModelProvider{
		ID:        id,
		API:       provider.API,
		URL:       provider.URL,
		KeyHint:   credentialHint(provider.Credential),
		Status:    provider.Status,
		CheckedAt: parseCheckedAt(provider.CheckedAt),
		Models:    provider.Models,
	}
}

func (c *Client) CreateModelProvider(ctx context.Context, name string, api values.ModelAPI, rawURL, key string) (*ModelProvider, error) {
	workspace, ws, err := c.callerWorkspaceRelease(ctx)

	if err != nil {
		return nil, err
	}

	id := ModelProviderID{Workspace: workspace, Name: name}

	if _, ok := ws.ModelProviders[name]; ok {
		return nil, usererror.Errorf("a model provider named %q already exists", name)
	}

	providerURL, err := modelProviderURL(api, rawURL)

	if err != nil {
		return nil, err
	}

	provider, err := c.checkModelProvider(ctx, api, providerURL, key)

	if err != nil {
		return nil, err
	}

	if err := c.deployer.Workspaces().SetModelProvider(ctx, id, provider); err != nil {
		return nil, err
	}

	return new(toModelProvider(id, provider)), nil
}

func (c *Client) ReplaceModelProviderKey(ctx context.Context, id ModelProviderID, key string) (*ModelProvider, error) {
	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	current, ok := ws.ModelProviders[id.Name]

	if !ok {
		return nil, usererror.Errorf("model provider %q not found", id.Name)
	}

	provider, err := c.checkModelProvider(ctx, current.API, current.URL, key)

	if err != nil {
		return nil, err
	}

	if err := c.deployer.Workspaces().SetModelProvider(ctx, id, provider); err != nil {
		return nil, err
	}

	return new(toModelProvider(id, provider)), nil
}

func (c *Client) RefreshModelProvider(ctx context.Context, id ModelProviderID) (*ModelProvider, error) {
	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	provider, ok := ws.ModelProviders[id.Name]

	if !ok {
		return nil, usererror.Errorf("model provider %q not found", id.Name)
	}

	models, providerURL, check := c.listModels(ctx, provider.API, provider.URL, provider.Credential)

	provider.Status = values.CredentialStatus(check.Status)
	provider.CheckedAt = time.Now().UTC().Format(time.RFC3339)

	if check.Status == upstream.Valid {
		provider.URL = providerURL
		provider.Models = models
	}

	if err := c.deployer.Workspaces().SetModelProvider(ctx, id, provider); err != nil {
		return nil, err
	}

	return new(toModelProvider(id, provider)), nil
}

func (c *Client) DeleteModelProvider(ctx context.Context, id ModelProviderID) (bool, error) {
	if err := c.deployer.Workspaces().DeleteModelProvider(ctx, id); err != nil {
		return false, err
	}

	return true, nil
}

// ModelProviderAgents returns the agents whose harness can use a provider.
func (c *Client) ModelProviderAgents(ctx context.Context, id ModelProviderID) ([]Agent, error) {
	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	provider, ok := ws.ModelProviders[id.Name]

	if !ok {
		return nil, usererror.Errorf("model provider %q not found", id.Name)
	}

	agents, err := c.workspaceAgents(ctx, id.Workspace, ws)

	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(agents, func(agent Agent) bool { return !harnessSupports(agent.Harness, provider.API) }), nil
}

func modelProviderURL(api values.ModelAPI, rawURL string) (string, error) {
	switch api {
	case values.ModelAPIAnthropic, values.ModelAPIOpenAI:
		fixed := anthropicURL

		if api == values.ModelAPIOpenAI {
			fixed = openAIURL
		}

		if rawURL != "" && strings.TrimSuffix(rawURL, "/") != fixed {
			return "", usererror.Errorf("the url of this API is fixed to %s", fixed)
		}

		return fixed, nil
	case values.ModelAPIOpenAICompatible:
		if rawURL == "" {
			return "", usererror.New("an OpenAI-compatible model provider needs a url")
		}

		return normalizeUpstreamURL(rawURL)
	default:
		return "", usererror.Errorf("unsupported api %q", api)
	}
}

func normalizeUpstreamURL(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))

	if err != nil {
		return "", usererror.New("url must be an https URL")
	}

	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	parsed.RawPath = ""

	normalized := parsed.String()

	if err := values.ValidateUpstreamURL(normalized); err != nil {
		return "", usererror.New(err.Error())
	}

	return normalized, nil
}

func (c *Client) checkModelProvider(ctx context.Context, api values.ModelAPI, providerURL, key string) (values.ModelProvider, error) {
	models, providerURL, check := c.listModels(ctx, api, providerURL, key)

	if check.Status != upstream.Valid {
		return values.ModelProvider{}, usererror.Errorf("couldn't connect the model provider: %s", check.Reason)
	}

	return values.ModelProvider{
		API:        api,
		URL:        providerURL,
		Credential: key,
		Status:     values.CredentialValid,
		CheckedAt:  time.Now().UTC().Format(time.RFC3339),
		Models:     models,
	}, nil
}

func (c *Client) listModels(ctx context.Context, api values.ModelAPI, providerURL, key string) ([]string, string, upstream.Check) {
	if api == values.ModelAPIAnthropic {
		models, check := c.upstream.AnthropicModels(ctx, providerURL, key)

		return models, providerURL, check
	}

	models, check := c.upstream.OpenAIModels(ctx, providerURL, key)

	if check.Status != upstream.Unreachable || api != values.ModelAPIOpenAICompatible || strings.HasSuffix(providerURL, "/v1") {
		return models, providerURL, check
	}

	alternative := providerURL + "/v1"

	if values.ValidateUpstreamURL(alternative) != nil {
		return models, providerURL, check
	}

	if retried, retriedCheck := c.upstream.OpenAIModels(ctx, alternative, key); retriedCheck.Status == upstream.Valid {
		return retried, alternative, retriedCheck
	}

	return models, providerURL, check
}

func credentialHint(credential string) string {
	if len(credential) < 8 {
		return ""
	}

	return credential[len(credential)-4:]
}

func parseCheckedAt(raw string) *time.Time {
	at, err := time.Parse(time.RFC3339, raw)

	if err != nil {
		return nil
	}

	return &at
}

func harnessSupports(harness values.Harness, api values.ModelAPI) bool {
	switch harness {
	case values.HarnessClaudeCode:
		return api == values.ModelAPIAnthropic
	case values.HarnessCodex:
		return api == values.ModelAPIOpenAI
	case values.HarnessOpenCode:
		return true
	default:
		return false
	}
}
