import { AgentStatus, Approval, Effort, Harness, ModelApi } from '@/gql/graphql';

export interface AgentConnectorGrant {
  connector: { id: string; name: string };
  approval: Approval;
  hiddenTools: string[];
}

export interface Agent {
  id: string;
  name: string;
  harness: Harness;
  instructions: string;
  status: AgentStatus;
  connectors: AgentConnectorGrant[];
  allowedHosts: string[];
  privateNetwork: boolean;
}

export interface HarnessOption {
  value: Harness;
  label: string;
  models: string;
}

export const harnesses: HarnessOption[] = [
  { value: Harness.ClaudeCode, label: 'Claude Code', models: 'Anthropic models' },
  { value: Harness.Codex, label: 'Codex', models: 'OpenAI models' },
  { value: Harness.Opencode, label: 'OpenCode', models: 'Models from every provider' },
];

export function harnessLabel(harness: Harness): string {
  return harnesses.find(option => option.value === harness)?.label ?? harness;
}

export const efforts: { value: Effort; label: string }[] = [
  { value: Effort.Low, label: 'Low' },
  { value: Effort.Medium, label: 'Medium' },
  { value: Effort.High, label: 'High' },
  { value: Effort.Max, label: 'Max' },
];

export const approvals: { value: Approval; label: string }[] = [
  { value: Approval.Allow, label: 'Allow' },
  { value: Approval.Ask, label: 'Ask first' },
];

export function modelApiLabel(api: ModelApi): string {
  switch (api) {
    case ModelApi.Anthropic:
      return 'Anthropic API';
    case ModelApi.Openai:
      return 'OpenAI API';
    default:
      return 'OpenAI-compatible';
  }
}

export interface ModelProviderPreset {
  id: string;
  label: string;
  name: string;
  api: ModelApi;
  url: string;
  urlPlaceholder?: string;
  keyLabel: string;
}

export const modelProviderPresets: ModelProviderPreset[] = [
  { id: 'anthropic', label: 'Anthropic', name: 'anthropic', api: ModelApi.Anthropic, url: '', keyLabel: 'API key' },
  { id: 'openai', label: 'OpenAI', name: 'openai', api: ModelApi.Openai, url: '', keyLabel: 'API key' },
  { id: 'mistral', label: 'Mistral', name: 'mistral', api: ModelApi.OpenaiCompatible, url: 'https://api.mistral.ai/v1', keyLabel: 'API key' },
  { id: 'infomaniak', label: 'Infomaniak', name: 'infomaniak', api: ModelApi.OpenaiCompatible, url: '', urlPlaceholder: 'https://api.infomaniak.com/2/ai/<product id>/openai/v1', keyLabel: 'API token' },
  { id: 'scaleway', label: 'Scaleway', name: 'scaleway', api: ModelApi.OpenaiCompatible, url: 'https://api.scaleway.ai/v1', keyLabel: 'Secret key' },
  { id: 'hetzner', label: 'Hetzner', name: 'hetzner', api: ModelApi.OpenaiCompatible, url: 'https://inference.hetzner.com/api/v1', keyLabel: 'API token' },
  { id: 'openai-compatible', label: 'OpenAI-compatible', name: '', api: ModelApi.OpenaiCompatible, url: '', urlPlaceholder: 'https://models.example.com/v1', keyLabel: 'API key' },
];

export interface ConnectorPreset {
  id: string;
  label: string;
  name: string;
  url: string;
}

export const connectorPresets: ConnectorPreset[] = [
  { id: 'github', label: 'GitHub', name: 'github', url: 'https://api.githubcopilot.com/mcp/' },
  { id: 'linear', label: 'Linear', name: 'linear', url: 'https://mcp.linear.app/mcp' },
  { id: 'mcp', label: 'Any MCP server', name: '', url: '' },
];

export interface AgentModelOption {
  provider: { id: string; name: string };
  name: string;
}

export function modelKey(model: AgentModelOption): string {
  return `${model.provider.id}\n${model.name}`;
}

export function missingModelsHint(harness: Harness): string {
  switch (harness) {
    case Harness.ClaudeCode:
      return 'Claude Code runs on Anthropic models, and the workspace has none yet.';
    case Harness.Codex:
      return 'Codex runs on OpenAI models, and the workspace has none yet.';
    default:
      return 'The workspace has no model provider yet.';
  }
}

export function resourceName(id: string): string {
  return id.split('/').pop() ?? id;
}
