import type { Component } from 'vue';
import { Container, GitCommitHorizontal, GitFork, Sprout } from '@lucide/vue';

import { graphql } from '@/gql';
import { Plan, SignalKind, type WorkspaceFieldsFragment } from '@/gql/graphql';
import { bytes, day, daysSince, daysUntil, duration, memory } from '@/lib/format';

export const WorkspaceFieldsDocument = graphql(`
  fragment WorkspaceFields on Workspace {
    id
    createdAt
    abuse
    abuseNote
    suspended
    billing { plan trialEndsAt goodStanding standingReason hasPaymentMethod }
    members { name email role account { login createdAt } }
    signals { kind strong detail }
    runningPods
    pendingPods
    memoryRequested
    egress7d
    environments {
      id
      project
      name
      runningPods
      pendingPods
      services { name image commit repository }
    }
  }
`);

export const WorkspacesDocument = graphql(`
  query Workspaces {
    workspaces { loadedAt workspaces { ...WorkspaceFields } }
  }
`);

export const CachedWorkspacesDocument = graphql(`
  query CachedWorkspaces {
    cachedWorkspaces { loadedAt workspaces { ...WorkspaceFields } }
  }
`);

export const MarkAbuseDocument = graphql(`
  mutation MarkAbuse($workspaceIDs: [ID!]!, $note: String) {
    markAbuse(workspaceIDs: $workspaceIDs, note: $note) { workspaceID error }
  }
`);

export const ClearAbuseDocument = graphql(`
  mutation ClearAbuse($workspaceIDs: [ID!]!) {
    clearAbuse(workspaceIDs: $workspaceIDs) { workspaceID error }
  }
`);

export type Workspace = WorkspaceFieldsFragment;
export type Signal = Workspace['signals'][number];

export type Outcome = {
  verb: 'marked' | 'cleared';
  results: { workspaceID: string; error?: string | null }[];
};

export type Tone = 'default' | 'muted' | 'warn' | 'danger' | 'progress';

export const toneClass: Record<Tone, string> = {
  default: '',
  muted: 'text-muted-foreground',
  warn: 'text-status-warn',
  danger: 'text-status-danger',
  progress: 'text-status-progress',
};

export function hasStrongSignal(workspace: Workspace): boolean {
  return workspace.signals.some((signal) => signal.strong);
}

export const signalIcons: Record<SignalKind, Component> = {
  [SignalKind.SharedCommit]: GitCommitHorizontal,
  [SignalKind.Image]: Container,
  [SignalKind.Fork]: GitFork,
  [SignalKind.YoungAccount]: Sprout,
};

export type FilterOption = {
  value: string;
  label: string;
  icon?: Component;
  test: (workspace: Workspace) => boolean;
};

export type ChoiceFilter = { key: string; label: string; options: FilterOption[] };

export type RangeFilter = {
  key: string;
  label: string;
  scale: 'log' | 'linear';
  integer?: boolean;
  value: (workspace: Workspace) => number | null;
  format: (value: number) => string;
};

export type Range = { min: number | null; max: number | null };

export type Filters = { choices: Record<string, string[]>; ranges: Record<string, Range> };

function signal(value: string, label: string, kind: SignalKind): FilterOption {
  return { value, label, icon: signalIcons[kind], test: (workspace) => workspace.signals.some((candidate) => candidate.kind === kind) };
}

function onPlan(value: string, label: string, plan: Plan): FilterOption {
  return { value, label, test: (workspace) => workspace.billing?.plan === plan };
}

export const choiceFilters: ChoiceFilter[] = [
  {
    key: 'plan',
    label: 'Plan',
    options: [
      { value: 'none', label: 'No subscription', test: (workspace) => workspace.billing == null },
      onPlan('trial', 'Trial', Plan.None),
      onPlan('hobby', 'Hobby', Plan.Hobby),
      onPlan('pro', 'Pro', Plan.Pro),
    ],
  },
  {
    key: 'signals',
    label: 'Signals',
    options: [
      signal('shared-commit', 'Shared commit', SignalKind.SharedCommit),
      signal('image', 'Image', SignalKind.Image),
      signal('fork', 'Fork', SignalKind.Fork),
      signal('young-account', 'Young account', SignalKind.YoungAccount),
      { value: 'none', label: 'No signals', test: (workspace) => workspace.signals.length === 0 },
    ],
  },
  {
    key: 'status',
    label: 'Status',
    options: [
      { value: 'active', label: 'Active', test: (workspace) => !workspace.abuse && !workspace.suspended },
      { value: 'suspending', label: 'Suspending', test: (workspace) => workspace.abuse && !workspace.suspended },
      { value: 'suspended', label: 'Suspended', test: (workspace) => workspace.suspended },
    ],
  },
];

export const rangeFilters: RangeFilter[] = [
  {
    key: 'owner-age',
    label: 'Owner age',
    scale: 'log',
    integer: true,
    value: (workspace) => {
      const account = owner(workspace)?.account;
      return account ? daysSince(account.createdAt) : null;
    },
    format: duration,
  },
  { key: 'egress', label: 'Egress 7d', scale: 'log', value: (workspace) => workspace.egress7d, format: bytes },
  {
    key: 'pods',
    label: 'Pods',
    scale: 'log',
    integer: true,
    value: (workspace) => workspace.runningPods + workspace.pendingPods,
    format: String,
  },
  { key: 'requests', label: 'Requests', scale: 'log', value: (workspace) => workspace.memoryRequested, format: memory },
  { key: 'created', label: 'Created', scale: 'linear', value: (workspace) => Date.parse(workspace.createdAt), format: day },
];

function inRange(value: number | null, range: Range | undefined): boolean {
  if (!range || (range.min == null && range.max == null)) return true;
  if (value == null) return false;
  return (range.min == null || value >= range.min) && (range.max == null || value <= range.max);
}

export function matches(workspace: Workspace, filters: Filters): boolean {
  const chosen = choiceFilters.every((filter) => {
    const values = filters.choices[filter.key] ?? [];
    return values.length === 0 || filter.options.some((option) => values.includes(option.value) && option.test(workspace));
  });

  return chosen && rangeFilters.every((filter) => inRange(filter.value(workspace), filters.ranges[filter.key]));
}

export function owner(workspace: Workspace): Workspace['members'][number] | undefined {
  return workspace.members.find((member) => member.role === 'admin') ?? workspace.members[0];
}

export function youngAccount(workspace: Workspace): boolean {
  return workspace.signals.some((signal) => signal.kind === SignalKind.YoungAccount);
}

export function signalLabel(signal: Signal): string {
  switch (signal.kind) {
    case SignalKind.SharedCommit:
      return `Commit ${signal.detail}`;
    case SignalKind.Image:
      return `Image ${signal.detail}`;
    case SignalKind.Fork:
      return `Fork of ${signal.detail}`;
    case SignalKind.YoungAccount:
      return `Account ${signal.detail}`;
  }
}

export function plan(workspace: Workspace): { text: string; tone: Tone } {
  const billing = workspace.billing;

  if (!billing) return { text: 'No subscription', tone: 'muted' };

  const plan = { [Plan.Pro]: 'Pro', [Plan.Hobby]: 'Hobby', [Plan.None]: 'Trial' }[billing.plan];

  if (!billing.goodStanding) {
    const reason = billing.standingReason ?? 'not in good standing';

    if (billing.plan === Plan.None && reason.startsWith('trial')) {
      return { text: reason.charAt(0).toUpperCase() + reason.slice(1), tone: 'warn' };
    }

    return { text: `${plan}, ${reason}`, tone: 'warn' };
  }

  if (billing.plan === Plan.None && billing.trialEndsAt) {
    const days = daysUntil(billing.trialEndsAt);
    return { text: `Trial, ${days} ${days === 1 ? 'day' : 'days'} left`, tone: 'default' };
  }

  return { text: plan, tone: 'default' };
}

export function status(workspace: Workspace): { text: string; note: string; tone: Tone } {
  if (workspace.suspended) {
    return { text: 'Suspended', note: workspace.abuse ? 'abuse' : (workspace.billing?.standingReason ?? ''), tone: 'muted' };
  }

  if (workspace.abuse) {
    return { text: 'Suspending', note: '', tone: 'progress' };
  }

  return { text: 'Active', note: '', tone: 'default' };
}

export function githubProfile(login: string): string {
  return `https://github.com/${encodeURIComponent(login)}`;
}

export async function openExternal(url: string) {
  await fetch('/open', { method: 'POST', body: url });
}

export function initial(id: string): string {
  return id.charAt(0).toUpperCase();
}

function hue(id: string): number {
  let sum = 0;
  for (const char of id) sum = (sum + char.charCodeAt(0)) % 10;
  return sum;
}

const hues: [string, string][] = [
  ['oklch(0.88 0.062 25)', 'oklch(0.42 0.12 25)'],
  ['oklch(0.88 0.072 55)', 'oklch(0.42 0.101 55)'],
  ['oklch(0.88 0.075 85)', 'oklch(0.42 0.083 85)'],
  ['oklch(0.88 0.075 125)', 'oklch(0.42 0.102 125)'],
  ['oklch(0.88 0.075 155)', 'oklch(0.42 0.101 155)'],
  ['oklch(0.88 0.075 190)', 'oklch(0.42 0.07 190)'],
  ['oklch(0.88 0.075 225)', 'oklch(0.42 0.077 225)'],
  ['oklch(0.88 0.057 260)', 'oklch(0.42 0.12 260)'],
  ['oklch(0.88 0.064 295)', 'oklch(0.42 0.12 295)'],
  ['oklch(0.88 0.075 340)', 'oklch(0.42 0.12 340)'],
];

export function avatarStyle(id: string): Record<string, string> {
  const [background, color] = hues[hue(id)]!;
  return { background, color };
}
