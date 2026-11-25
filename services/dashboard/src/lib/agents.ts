import { AgentStatus, AgentTriggerSource, Approval, ChatAppKind, Effort, Harness, ModelApi } from '@/gql/graphql';

export interface AgentConnectorGrant {
  connector: { id: string; name: string };
  approval: Approval;
  hiddenTools: string[];
}

export interface AgentChat {
  chatApp: { id: string; name: string; kind: ChatAppKind; account: string; url?: string | null };
  channel: { id: string; name: string };
}

export interface Agent {
  id: string;
  name: string;
  harness: Harness;
  instructions: string;
  status: AgentStatus;
  connectors: AgentConnectorGrant[];
  chat?: AgentChat | null;
  triggers: AgentTrigger[];
  allowedHosts: string[];
  privateNetwork: boolean;
}

export interface AgentTrigger {
  id: string;
  name: string;
  source: AgentTriggerSource;
  schedule?: string | null;
  timeZone?: string | null;
  key?: string | null;
  instructions: string;
  enabled: boolean;
  url?: string | null;
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

export type ModelProviderBrand = 'anthropic' | 'openai' | 'mistral' | 'infomaniak' | 'scaleway' | 'hetzner' | 'opencode' | 'generic';

export interface ModelProviderPreset {
  id: string;
  label: string;
  name: string;
  brand: ModelProviderBrand;
  api: ModelApi;
  url: string;
  domain?: string;
  urlPlaceholder?: string;
  urlHint?: string;
  keyLabel: string;
  keyUrl?: string;
  keyPlaceholder?: string;
  keyHint?: string;
}

export const modelProviderPresets: ModelProviderPreset[] = [
  {
    id: 'anthropic',
    label: 'Anthropic',
    name: 'anthropic',
    brand: 'anthropic',
    api: ModelApi.Anthropic,
    url: '',
    keyLabel: 'API key',
    keyUrl: 'https://console.anthropic.com/settings/keys',
    keyPlaceholder: 'sk-ant-…',
  },
  {
    id: 'openai',
    label: 'OpenAI',
    name: 'openai',
    brand: 'openai',
    api: ModelApi.Openai,
    url: '',
    keyLabel: 'API key',
    keyUrl: 'https://platform.openai.com/api-keys',
    keyPlaceholder: 'sk-…',
  },
  {
    id: 'mistral',
    label: 'Mistral',
    name: 'mistral',
    brand: 'mistral',
    api: ModelApi.OpenaiCompatible,
    url: 'https://api.mistral.ai/v1',
    domain: 'mistral.ai',
    keyLabel: 'API key',
    keyUrl: 'https://console.mistral.ai/home?profile_dialog=api-keys',
  },
  {
    id: 'infomaniak',
    label: 'Infomaniak',
    name: 'infomaniak',
    brand: 'infomaniak',
    api: ModelApi.OpenaiCompatible,
    url: '',
    domain: 'infomaniak.com',
    urlPlaceholder: 'https://api.infomaniak.com/2/ai/<product id>/openai/v1',
    keyLabel: 'API token',
  },
  {
    id: 'scaleway',
    label: 'Scaleway',
    name: 'scaleway',
    brand: 'scaleway',
    api: ModelApi.OpenaiCompatible,
    url: 'https://api.scaleway.ai/v1',
    domain: 'scaleway.ai',
    keyLabel: 'Secret key',
    keyUrl: 'https://console.scaleway.com/iam/api-keys',
  },
  {
    id: 'hetzner',
    label: 'Hetzner',
    name: 'hetzner',
    brand: 'hetzner',
    api: ModelApi.OpenaiCompatible,
    url: 'https://inference.hetzner.com/api/v1',
    domain: 'hetzner.com',
    keyLabel: 'API token',
    keyUrl: 'https://experiments.hetzner.com/inference',
  },
  {
    id: 'opencode-go',
    label: 'OpenCode Go',
    name: 'opencode-go',
    brand: 'opencode',
    api: ModelApi.OpenaiCompatible,
    url: 'https://opencode.ai/zen/go/v1',
    domain: 'opencode.ai',
    keyLabel: 'API key',
    keyUrl: 'https://opencode.ai/auth',
    keyHint: 'From the OpenCode console, once you subscribe to Go. OpenCode lists its models without a key, so a wrong key is not caught here.',
  },
  {
    id: 'custom',
    label: 'Custom',
    name: '',
    brand: 'generic',
    api: ModelApi.OpenaiCompatible,
    url: '',
    urlPlaceholder: 'https://models.example.com/v1',
    urlHint: 'The base URL of any OpenAI-compatible API, the one that serves /models and /chat/completions.',
    keyLabel: 'API key',
  },
];

export interface ModelProviderIdentity {
  name: string;
  api?: ModelApi | null;
  url?: string | null;
}

function hostname(url: string): string | null {
  try {
    return new URL(url).hostname.toLowerCase();
  } catch {
    return null;
  }
}

export function modelProviderBrand(provider: ModelProviderIdentity): ModelProviderBrand {
  if (provider.api === ModelApi.Anthropic) return 'anthropic';
  if (provider.api === ModelApi.Openai) return 'openai';
  const host = provider.url ? hostname(provider.url) : null;
  if (host) {
    const preset = modelProviderPresets.find(candidate => candidate.domain && (host === candidate.domain || host.endsWith(`.${candidate.domain}`)));
    return preset?.brand ?? 'generic';
  }
  if (provider.api) return 'generic';
  return modelProviderPresets.find(candidate => candidate.name && candidate.name === provider.name.trim().toLowerCase())?.brand ?? 'generic';
}

export interface ConnectorPreset {
  id: string;
  label: string;
  name: string;
  url: string;
  keyLabel: string;
  keyUrl?: string;
}

export const connectorPresets: ConnectorPreset[] = [
  { id: 'github', label: 'GitHub', name: 'github', url: 'https://api.githubcopilot.com/mcp/', keyLabel: 'Personal access token', keyUrl: 'https://github.com/settings/personal-access-tokens' },
  { id: 'linear', label: 'Linear', name: 'linear', url: 'https://mcp.linear.app/mcp', keyLabel: 'API key', keyUrl: 'https://linear.app/settings/account/security' },
  { id: 'mcp', label: 'Any MCP server', name: '', url: '', keyLabel: 'Token' },
];

export const chatAppKinds: { value: ChatAppKind; label: string }[] = [
  { value: ChatAppKind.Telegram, label: 'Telegram' },
  { value: ChatAppKind.Mattermost, label: 'Mattermost' },
];

export function chatAppKindLabel(kind: ChatAppKind): string {
  return chatAppKinds.find(option => option.value === kind)?.label ?? kind;
}

export interface AgentModelOption {
  provider: { id: string; name: string; api?: ModelApi | null; url?: string | null };
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
