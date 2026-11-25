import { AgentStatus, AgentTriggerSource, Approval, ChatAppKind, Effort, Harness, ModelApi } from '@/gql/graphql';

export interface AgentConnectorGrant {
  connector: { id: string; name: string; url: string };
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
  repository?: string | null;
  events?: string[] | null;
  branches?: string[] | null;
  onlyFailures?: boolean | null;
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

export interface ConnectorToolset {
  id: string;
  label: string;
  url: string;
  hint: string;
}

export interface ConnectorPreset {
  id: string;
  label: string;
  name: string;
  url: string;
  domain?: string;
  urlPlaceholder?: string;
  toolsets?: ConnectorToolset[];
  keyLabel: string;
  keyUrl?: string;
  keyHint?: string;
  keyOptional?: boolean;
}

const githubToolsets: ConnectorToolset[] = [
  { id: 'default', label: 'Default tools', url: 'https://api.githubcopilot.com/mcp/', hint: 'Repositories, issues, pull requests and users, with the tools that write.' },
  { id: 'default-readonly', label: 'Default tools, read only', url: 'https://api.githubcopilot.com/mcp/readonly', hint: 'The default tools without any that write, including ones GitHub adds later.' },
  { id: 'all', label: 'All tools', url: 'https://api.githubcopilot.com/mcp/x/all', hint: 'Every toolset, Actions and security included. Hide what an agent shouldn\'t use.' },
  { id: 'all-readonly', label: 'All tools, read only', url: 'https://api.githubcopilot.com/mcp/x/all/readonly', hint: 'Every toolset without any tool that writes, including ones GitHub adds later.' },
];

export const connectorPresets: ConnectorPreset[] = [
  { id: 'github', label: 'GitHub', name: 'github', domain: 'githubcopilot.com', url: githubToolsets[0]!.url, toolsets: githubToolsets, keyLabel: 'Personal access token', keyUrl: 'https://github.com/settings/personal-access-tokens' },
  { id: 'gitlab', label: 'GitLab', name: 'gitlab', domain: 'gitlab.com', url: 'https://gitlab.com/api/v4/mcp', urlPlaceholder: 'https://gitlab.example.com/api/v4/mcp', keyLabel: 'Personal access token', keyUrl: 'https://gitlab.com/-/user_settings/personal_access_tokens', keyHint: 'A fine-grained token with the MCP permission.' },
  { id: 'linear', label: 'Linear', name: 'linear', domain: 'linear.app', url: 'https://mcp.linear.app/mcp', keyLabel: 'API key', keyUrl: 'https://linear.app/settings/account/security' },
  { id: 'atlassian', label: 'Atlassian', name: 'atlassian', domain: 'atlassian.com', url: 'https://mcp.atlassian.com/v2/mcp', keyLabel: 'Service account API key', keyUrl: 'https://admin.atlassian.com', keyHint: 'An org admin creates it. Personal API tokens don\'t work here.' },
  { id: 'notion', label: 'Notion', name: 'notion', domain: 'notion.com', url: 'https://mcp.notion.com/mcp', keyLabel: 'MCP token', keyHint: 'Workspace owners on Business or Enterprise create it.' },
  { id: 'airtable', label: 'Airtable', name: 'airtable', domain: 'airtable.com', url: 'https://mcp.airtable.com/mcp', keyLabel: 'Personal access token', keyUrl: 'https://airtable.com/create/tokens' },
  { id: 'monday', label: 'monday.com', name: 'monday', domain: 'monday.com', url: 'https://mcp.monday.com/mcp', keyLabel: 'Personal API token' },
  { id: 'odoo', label: 'Odoo', name: 'odoo', domain: 'odoo.com', url: '', urlPlaceholder: 'https://mycompany.odoo.com/mcp', keyLabel: 'API key', keyHint: 'Scoped to MCP. Needs Odoo 20 or a current Odoo Online.' },
  { id: 'stripe', label: 'Stripe', name: 'stripe', domain: 'stripe.com', url: 'https://mcp.stripe.com', keyLabel: 'Restricted key', keyUrl: 'https://dashboard.stripe.com/apikeys', keyHint: 'A restricted key tagged for agents.' },
  { id: 'brevo', label: 'Brevo', name: 'brevo', domain: 'brevo.com', url: 'https://mcp.brevo.com/v1/brevo/mcp', keyLabel: 'MCP server API key' },
  { id: 'zapier', label: 'Zapier', name: 'zapier', domain: 'zapier.com', url: 'https://mcp.zapier.com/api/v1/connect', keyLabel: 'Connection token', keyUrl: 'https://mcp.zapier.com' },
  { id: 'make', label: 'Make', name: 'make', domain: 'make.com', url: '', urlPlaceholder: 'https://eu2.make.com/mcp/v2', keyLabel: 'API token', keyHint: 'With the mcp:use scope. The URL starts with your zone.' },
  { id: 'n8n', label: 'n8n', name: 'n8n', domain: 'n8n.cloud', url: '', urlPlaceholder: 'https://n8n.example.com/mcp-server/http', keyLabel: 'MCP access token' },
  { id: 'kubernetes', label: 'Kubernetes', name: 'kubernetes', url: '', urlPlaceholder: 'https://kubernetes-mcp.example.com/mcp', keyLabel: 'Token', keyOptional: true },
  { id: 'mcp', label: 'Any MCP server', name: '', url: '', urlPlaceholder: 'https://mcp.example.com/mcp', keyLabel: 'Token', keyOptional: true },
];

export function connectorBrand(connector: { name: string; url?: string | null }): string | null {
  const host = connector.url ? hostname(connector.url) : null;
  const preset = host ? connectorPresets.find(candidate => candidate.domain && (host === candidate.domain || host.endsWith(`.${candidate.domain}`))) : undefined;
  if (preset) return preset.name;
  if (`${connector.name} ${connector.url ?? ''}`.toLowerCase().includes('kubernetes')) return 'kubernetes';
  return null;
}

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
