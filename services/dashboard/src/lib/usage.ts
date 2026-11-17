import type { RouteLocationNamedRaw } from 'vue-router';
import type { PanelEntry } from '@/composables/usePanel';
import { InvoiceLineKind, UsageResourceKind } from '@/gql/graphql';
import { formatMoney } from '@/lib/format';
import { environmentRoute, projectRoute } from '@/lib/routes';

const LOCALE = 'en-US';
const DAY = 24 * 60 * 60 * 1000;

const palette = ['var(--chart-1)', 'var(--chart-5)', 'var(--chart-4)', 'var(--chart-3)', 'var(--chart-2)'];
const neutral = 'var(--muted-foreground)';
const shades = [100, 60, 36, 20];
const resourceShades = [70, 45, 28, 16];

const BREAKDOWN_SHARE = 0.95;
const BREAKDOWN_SLACK_CENTS = 100;

export const invoiceLineLabels: Record<InvoiceLineKind, string> = {
  [InvoiceLineKind.Plan]: 'Plan',
  [InvoiceLineKind.Cpu]: 'CPU',
  [InvoiceLineKind.Memory]: 'Memory',
  [InvoiceLineKind.Volumes]: 'Volumes',
  [InvoiceLineKind.ObjectStorage]: 'Object storage',
  [InvoiceLineKind.Egress]: 'Egress',
};

export function breakdownComplete(breakdownCents: number, invoicedCents: number): boolean {
  const missing = invoicedCents - breakdownCents;
  return missing < BREAKDOWN_SLACK_CENTS || breakdownCents >= invoicedCents * BREAKDOWN_SHARE;
}

const panelTypes: Record<UsageResourceKind, PanelEntry['type']> = {
  [UsageResourceKind.Service]: 'service',
  [UsageResourceKind.Database]: 'database',
  [UsageResourceKind.KeyValueStore]: 'keyValueStore',
  [UsageResourceKind.Bucket]: 'bucket',
  [UsageResourceKind.Volume]: 'volume',
  [UsageResourceKind.Agent]: 'agent',
};
const FOLD_SHARE = 0.01;

export interface UsageResource {
  resourceId: string;
  name: string;
  kind: UsageResourceKind;
  amountCents: number;
  daily: number[];
}

export interface UsageEnvironment {
  environmentId: string;
  name: string;
  deletedAt?: string | null;
  amountCents: number;
  daily: number[];
  resources: UsageResource[];
}

export interface UsageProject {
  projectId: string;
  name: string;
  deletedAt?: string | null;
  amountCents: number;
  daily: number[];
  environments: UsageEnvironment[];
}

export interface UsageScope {
  project: string;
  environment?: string;
}

export interface UsageRow {
  key: string;
  name: string;
  kind?: UsageResourceKind;
  deletedAt?: string | null;
  amountCents: number;
  daily: number[];
  color?: string;
  scope?: UsageScope;
  to?: RouteLocationNamedRaw;
  panel?: PanelEntry;
  children: UsageRow[];
}

export type UsageLevel = 'project' | 'environment' | 'resource';

export function tint(color: string, amount: number): string {
  return amount >= 100 ? color : `color-mix(in oklab, ${color} ${amount}%, var(--card))`;
}

export function projectColors(projects: UsageProject[]): Map<string, string> {
  const colors = new Map<string, string>();

  projects.forEach((project, i) => {
    const color = palette[i] ?? neutral;
    colors.set(project.projectId, color);

    project.environments.forEach((environment, j) => {
      colors.set(environment.environmentId, tint(color, shades[Math.min(j, shades.length - 1)]!));
    });
  });

  return colors;
}

export function formatCents(cents: number): string {
  if (cents > 0 && cents < 0.5) return `<${formatMoney(0.01, 'EUR')}`;
  return formatMoney(cents / 100, 'EUR');
}

export function formatRate(cents: number): string {
  const euros = cents / 100;
  const options: Intl.NumberFormatOptions = Math.abs(euros) >= 1
    ? { style: 'currency', currency: 'EUR' }
    : { style: 'currency', currency: 'EUR', maximumSignificantDigits: 4 };
  return new Intl.NumberFormat(LOCALE, options).format(euros).replace('-', '−');
}

export function formatQuantity(value: number): string {
  if (value > 0 && value < 0.01) return '<0.01';
  const digits = value < 10 ? 2 : value < 1000 ? 1 : 0;
  return value.toLocaleString(LOCALE, { maximumFractionDigits: digits });
}

export function formatShare(ratio: number): string {
  if (ratio <= 0) return '0%';
  if (ratio < 0.01) return '<1%';
  return `${Math.round(ratio * 100)}%`;
}

export function formatDay(date: Date): string {
  return date.toLocaleDateString(LOCALE, { month: 'short', day: 'numeric' });
}

export function formatMonth(date: Date): string {
  return date.toLocaleDateString(LOCALE, { month: 'long' });
}

export function formatDayTime(date: Date): string {
  return date.toLocaleString(LOCALE, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', hourCycle: 'h23' });
}

export function dayDate(periodStart: string, index: number): Date {
  return new Date(new Date(periodStart).getTime() + index * DAY);
}

export function dayIndex(periodStart: string, at: string): number {
  return Math.floor((new Date(at).getTime() - new Date(periodStart).getTime()) / DAY);
}

export function periodLabel(start: string, end: string): string {
  return `${formatDay(new Date(start))} to ${formatDay(new Date(new Date(end).getTime() - 1))}`;
}

export function scopeKey(scope: UsageScope | null): string {
  return scope ? scope.environment ?? scope.project : '';
}

export function scopeLevel(scope: UsageScope | null): UsageLevel {
  if (!scope) return 'project';
  return scope.environment ? 'resource' : 'environment';
}

function resourceRow(resource: UsageResource, environment: UsageEnvironment, color?: string): UsageRow {
  return {
    key: `${resource.kind}:${resource.resourceId}`,
    name: resource.name,
    kind: resource.kind,
    amountCents: resource.amountCents,
    daily: resource.daily,
    color,
    to: environment.deletedAt ? undefined : environmentRoute(environment.environmentId),
    panel: { type: panelTypes[resource.kind], id: resource.resourceId, label: resource.name },
    children: [],
  };
}

function environmentRow(project: UsageProject, environment: UsageEnvironment, colors: Map<string, string>): UsageRow {
  return {
    key: environment.environmentId,
    name: environment.name,
    deletedAt: environment.deletedAt,
    amountCents: environment.amountCents,
    daily: environment.daily,
    color: colors.get(environment.environmentId),
    scope: { project: project.projectId, environment: environment.environmentId },
    to: environment.deletedAt ? undefined : environmentRoute(environment.environmentId),
    children: environment.resources.map(resource => resourceRow(resource, environment)),
  };
}

export function usageRows(projects: UsageProject[], scope: UsageScope | null, colors: Map<string, string>): UsageRow[] {
  if (!scope) {
    return projects.map(project => ({
      key: project.projectId,
      name: project.name,
      deletedAt: project.deletedAt,
      amountCents: project.amountCents,
      daily: project.daily,
      color: colors.get(project.projectId),
      scope: { project: project.projectId },
      to: project.deletedAt ? undefined : projectRoute(project.projectId),
      children: project.environments.map(environment => environmentRow(project, environment, colors)),
    }));
  }

  const project = projects.find(candidate => candidate.projectId === scope.project);
  if (!project) return [];

  if (!scope.environment) {
    return project.environments.map(environment => environmentRow(project, environment, colors));
  }

  const environment = project.environments.find(candidate => candidate.environmentId === scope.environment);
  if (!environment) return [];

  const color = colors.get(project.projectId) ?? neutral;

  return environment.resources.map((resource, i) =>
    resourceRow(resource, environment, tint(color, resourceShades[Math.min(i, resourceShades.length - 1)]!)));
}

export function foldRows(rows: UsageRow[]): { main: UsageRow[]; tail: UsageRow[] } {
  const total = sum(rows.map(row => row.amountCents));
  const tail = rows.filter(row => total > 0 && row.amountCents / total < FOLD_SHARE);
  if (tail.length < 2) return { main: rows, tail: [] };
  return { main: rows.filter(row => !tail.includes(row)), tail };
}

export function foldRow(tail: UsageRow[], level: UsageLevel): UsageRow {
  return {
    key: 'tail',
    name: `${tail.length} ${level}s under ${formatShare(FOLD_SHARE)}`,
    amountCents: sum(tail.map(row => row.amountCents)),
    daily: sumDays(tail.map(row => row.daily)),
    color: tint(neutral, 45),
    children: [],
  };
}

export function scoped(projects: UsageProject[], scope: UsageScope | null): UsageProject[] {
  if (!scope) return projects;

  return projects
    .filter(project => project.projectId === scope.project)
    .map((project) => {
      if (!scope.environment) return project;
      const environments = project.environments.filter(environment => environment.environmentId === scope.environment);
      return {
        ...project,
        amountCents: sum(environments.map(environment => environment.amountCents)),
        daily: sumDays(environments.map(environment => environment.daily)),
        environments,
      };
    });
}

export function sum(values: number[]): number {
  return values.reduce((total, value) => total + value, 0);
}

export function sumDays(lists: number[][]): number[] {
  const length = Math.max(0, ...lists.map(list => list.length));
  return Array.from({ length }, (_, i) => sum(lists.map(list => list[i] ?? 0)));
}

const kindLabels: Record<UsageResourceKind, string> = {
  [UsageResourceKind.Service]: 'service',
  [UsageResourceKind.Database]: 'database',
  [UsageResourceKind.KeyValueStore]: 'key-value store',
  [UsageResourceKind.Volume]: 'volume',
  [UsageResourceKind.Bucket]: 'bucket',
  [UsageResourceKind.Agent]: 'agent',
};

function csvField(value: string): string {
  return /[",\n]/.test(value) ? `"${value.replace(/"/g, '""')}"` : value;
}

export function isoDay(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

export function usageCsv(projects: UsageProject[], periodStart: string): string {
  const rows = [['date', 'project', 'environment', 'resource', 'kind', 'amount_eur']];

  for (const project of projects) {
    for (const environment of project.environments) {
      environment.daily.forEach((cents, i) => {
        if (cents <= 0) return;

        const date = isoDay(dayDate(periodStart, i));
        let rest = cents;

        for (const resource of environment.resources) {
          const amount = resource.daily[i] ?? 0;
          if (amount <= 0) continue;
          rest -= amount;
          rows.push([date, project.name, environment.name, resource.name, kindLabels[resource.kind], (amount / 100).toFixed(4)]);
        }

        if (rest >= 0.005) {
          rows.push([date, project.name, environment.name, '', '', (rest / 100).toFixed(4)]);
        }
      });
    }
  }

  return rows.map(row => row.map(csvField).join(',')).join('\n') + '\n';
}

export function download(filename: string, content: string, type: string) {
  const url = URL.createObjectURL(new Blob([content], { type }));
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}
