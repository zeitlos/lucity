<script setup lang="ts">
import { environmentRoute, projectRoute, routeEnvironmentId } from '@/lib/routes';
import { computed, nextTick, ref, watch, type Component, type StyleValue } from 'vue';
import { useRoute, useRouter, type RouteLocationRaw } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { useEventListener } from '@vueuse/core';
import {
  ArrowUpRight,
  BookOpen,
  Braces,
  ChartLine,
  Check,
  ChevronRight,
  DatabaseBackup,
  FolderOpen,
  Gauge,
  History,
  Layers,
  LogOut,
  MessageSquarePlus,
  Moon,
  Plug,
  Plus,
  Rocket,
  ScrollText,
  Search,
  ShieldCheck,
  SlidersHorizontal,
  SquareTerminal,
  Sun,
  Table2,
} from '@lucide/vue';
import { graphql } from '@/gql';
import { EndpointType } from '@/gql/graphql';
import IconButton from '@/components/IconButton.vue';
import ResourceIcon from '@/components/ResourceIcon.vue';
import WorkspaceSwitchDialog from '@/components/WorkspaceSwitchDialog.vue';
import { Skeleton } from '@/components/ui/skeleton';
import { useAuth } from '@/composables/useAuth';
import { useCapabilities } from '@/composables/useCapabilities';
import { useCommandPalette } from '@/composables/useCommandPalette';
import { useContextSwitch } from '@/composables/useContextSwitch';
import { useDeploy } from '@/composables/useDeploy';
import { useEnvironment } from '@/composables/useEnvironment';
import { OPEN_LAYER, isEditable } from '@/composables/useEscapeToClose';
import { usePanel, type PanelEntry } from '@/composables/usePanel';
import { useReportBug } from '@/composables/useReportBug';
import { useServiceLogsPanel } from '@/composables/useServiceLogsPanel';
import { useTheme } from '@/composables/useTheme';
import { useWorkspaceColor } from '@/composables/useWorkspaceColor';
import { workspaceNavItems } from '@/lib/contextNav';
import { recentResources, rememberResource } from '@/lib/recentResources';
import { serviceKind, type ResourceKind } from '@/lib/resources';

const CommandPaletteProjectsDocument = graphql(`
  query CommandPaletteProjects {
    projects {
      id
      name
      environments {
        id
        name
        services {
          id
          name
          sourceUrl
        }
        databases {
          id
          name
        }
        keyValueStores {
          id
          name
        }
        buckets {
          id
          name
        }
        volumes {
          id
          name
        }
      }
    }
  }
`);

type ResourceType = Exclude<PanelEntry['type'], 'agent'>;

interface PaletteProject {
  id: string;
  name: string;
}

interface PaletteEnvironment {
  id: string;
  name: string;
  project: PaletteProject;
}

interface PaletteResource {
  id: string;
  name: string;
  type: ResourceType;
  kind: ResourceKind;
  environment: PaletteEnvironment;
}

interface ScopeEntry {
  id: string;
  name: string;
}

interface Scope {
  project?: ScopeEntry;
  environment?: ScopeEntry;
  resource?: ScopeEntry;
}

interface PaletteItem {
  key: string;
  label: string;
  group: string;
  hint: [string, string?];
  icon?: Component;
  resourceKind?: ResourceKind;
  letter?: { text: string; class: string; style?: StyleValue };
  meta?: string;
  keywords?: string;
  matchLabel?: string;
  boost?: number;
  shortcut?: string;
  current?: boolean;
  to?: RouteLocationRaw;
  enter?: Scope;
  run: () => unknown;
}

const DOCS_URL = 'https://lucity.cloud/docs/quickstart';

const GROUP_ORDER = ['elsewhere', 'recent', 'resources','environments', 'projects', 'tabs', 'actions', 'pages', 'workspaces', 'app', 'new'];

const TABS: Record<ResourceType, { key: string; label: string; icon: Component }[]> = {
  service: [
    { key: 'deployments', label: 'Deployments', icon: History },
    { key: 'logs', label: 'Logs', icon: ScrollText },
    { key: 'metrics', label: 'Metrics', icon: ChartLine },
    { key: 'variables', label: 'Variables', icon: Braces },
    { key: 'security', label: 'Security', icon: ShieldCheck },
    { key: 'settings', label: 'Settings', icon: SlidersHorizontal },
  ],
  database: [
    { key: 'connect', label: 'Connect', icon: Plug },
    { key: 'metrics', label: 'Metrics', icon: ChartLine },
    { key: 'tables', label: 'Tables', icon: Table2 },
    { key: 'query', label: 'Query', icon: SquareTerminal },
    { key: 'backups', label: 'Backups', icon: DatabaseBackup },
    { key: 'settings', label: 'Settings', icon: SlidersHorizontal },
  ],
  keyValueStore: [
    { key: 'connect', label: 'Connect', icon: Plug },
    { key: 'settings', label: 'Settings', icon: SlidersHorizontal },
  ],
  bucket: [
    { key: 'files', label: 'Files', icon: FolderOpen },
    { key: 'connect', label: 'Connect', icon: Plug },
    { key: 'settings', label: 'Settings', icon: SlidersHorizontal },
  ],
  volume: [
    { key: 'usage', label: 'Usage', icon: Gauge },
    { key: 'settings', label: 'Settings', icon: SlidersHorizontal },
  ],
};

const route = useRoute();
const router = useRouter();
const { open, modKey } = useCommandPalette();

const helpRows = [
  { label: 'Open the focused row', key: '↵' },
  { label: 'Open it in a new tab', key: `${modKey}↵` },
  { label: 'Go into it', key: '→' },
  { label: 'Go up one level', key: '⌫' },
  { label: 'Close', key: 'esc' },
  { label: 'Open the palette', key: `${modKey}K` },
  { label: 'New', key: 'N' },
];

const { activeWorkspace, isWorkspaceAdmin, logout } = useAuth();
const { agents } = useCapabilities();
const {
  workspaces,
  activeWs,
  activeProjectId,
  activeProjectName,
  workspaceSwitch,
  handleWorkspaceSwitch,
  finishWorkspaceSwitch,
} = useContextSwitch();
const { activeEnvironment, activeEnvServices } = useEnvironment();
const { currentPanel, openPanel } = usePanel();
const serviceLogsPanel = useServiceLogsPanel();
const { requestFeature } = useReportBug();
const { theme, toggleTheme } = useTheme();
const { workspaceColor } = useWorkspaceColor();

const deployTarget = ref<PaletteResource | null>(null);
const deploy = useDeploy(() => ({
  id: deployTarget.value?.id ?? '',
  name: deployTarget.value?.name ?? '',
  builds: [],
}));

const { result, loading } = useQuery(CommandPaletteProjectsDocument, null, () => ({
  enabled: open.value,
  fetchPolicy: 'cache-and-network',
}));

const index = computed(() => {
  const projects: PaletteProject[] = [];
  const environments: PaletteEnvironment[] = [];
  const resources: PaletteResource[] = [];

  for (const p of result.value?.projects ?? []) {
    const project = { id: p.id, name: p.name };
    projects.push(project);

    for (const e of p.environments) {
      const environment = { id: e.id, name: e.name, project };
      environments.push(environment);
      resources.push(
        ...e.services.map(s => ({ id: s.id, name: s.name, type: 'service' as const, kind: serviceKind(s.sourceUrl), environment })),
        ...e.databases.map(d => ({ id: d.id, name: d.name, type: 'database' as const, kind: 'database' as const, environment })),
        ...e.keyValueStores.map(k => ({ id: k.id, name: k.name, type: 'keyValueStore' as const, kind: 'keyValueStore' as const, environment })),
        ...e.buckets.map(b => ({ id: b.id, name: b.name, type: 'bucket' as const, kind: 'bucket' as const, environment })),
        ...e.volumes.map(v => ({ id: v.id, name: v.name, type: 'volume' as const, kind: 'volume' as const, environment })),
      );
    }
  }

  return { projects, environments, resources };
});

const scope = ref<Scope>({});
const query = ref('');
const focusedIndex = ref(0);
const recentIds = ref<string[]>([]);
const inputRef = ref<HTMLInputElement | null>(null);
const listRef = ref<HTMLElement | null>(null);

const scopeProject = computed(() => index.value.projects.find(p => p.id === scope.value.project?.id));
const scopeEnvironment = computed(() => index.value.environments.find(e => e.id === scope.value.environment?.id));
const scopeResource = computed(() => index.value.resources.find(r => r.id === scope.value.resource?.id));

const path = computed(() => {
  const { project, environment, resource } = scope.value;
  const segments: { id: string; name: string; kind?: ResourceKind }[] = [];
  if (project) segments.push(project);
  if (environment) segments.push(environment);
  if (resource) segments.push({ ...resource, kind: scopeResource.value?.kind });
  return segments;
});

const upHint = computed(() => {
  if (query.value) return null;
  if (scope.value.resource) return `Up to ${scope.value.environment?.name}`;
  if (scope.value.environment) return `Up to ${scope.value.project?.name}`;
  if (scope.value.project) return `Up to ${activeWs.value?.name ?? 'all projects'}`;
  return null;
});

function entry(item: { id: string; name: string }): ScopeEntry {
  return { id: item.id, name: item.name };
}

function scopeOfEnvironment(environment: PaletteEnvironment): Scope {
  return { project: entry(environment.project), environment: entry(environment) };
}

function scopeOfResource(resource: PaletteResource): Scope {
  return { ...scopeOfEnvironment(resource.environment), resource: entry(resource) };
}

function scopeFromRoute(): Scope {
  if (!activeProjectId.value) return {};
  const project = { id: activeProjectId.value, name: activeProjectName.value };
  if (route.name !== 'environment') return route.name === 'project' ? { project } : {};

  const environmentId = routeEnvironmentId(route) ?? '';
  const known = index.value.environments.find(e => e.id === environmentId);
  const environment = { id: environmentId, name: known?.name ?? activeEnvironment.value?.name ?? environmentId.split('/').pop() ?? '' };
  const panel = currentPanel.value;
  if (!panel || panel.type === 'agent') return { project, environment };
  return { project, environment, resource: { id: panel.id, name: panel.label } };
}

function environmentLocation(environment: PaletteEnvironment, drawer?: string) {
  return environmentRoute(environment.id, drawer);
}

function location(environment: PaletteEnvironment) {
  if (environment.id === scope.value.environment?.id) return undefined;
  if (environment.project.id === scope.value.project?.id) return environment.name;
  return `${environment.project.name} / ${environment.name}`;
}

async function openResource(resource: PaletteResource, tab?: string) {
  await router.push(environmentLocation(resource.environment));
  await openPanel({ type: resource.type, id: resource.id, label: resource.name, tab: tab === 'logs' ? undefined : tab });
  if (tab !== 'logs' || currentPanel.value?.id !== resource.id) return;
  await nextTick();
  serviceLogsPanel.open(resource.id, resource.name);
}

function resourceItem(resource: PaletteResource, group = 'resources'): PaletteItem {
  return {
    key: `${group}:${resource.id}`,
    label: resource.name,
    group,
    hint: ['Open', resource.name],
    resourceKind: resource.kind,
    meta: location(resource.environment),
    enter: scopeOfResource(resource),
    run: () => openResource(resource),
  };
}

function environmentItem(environment: PaletteEnvironment): PaletteItem {
  const to = environmentLocation(environment);
  return {
    key: `environment:${environment.id}`,
    label: environment.name,
    group: 'environments',
    hint: ['Open', environment.name],
    icon: Layers,
    meta: environment.project.id === scope.value.project?.id ? undefined : environment.project.name,
    to,
    enter: scopeOfEnvironment(environment),
    run: () => router.push(to),
  };
}

function projectItem(project: PaletteProject): PaletteItem {
  const to = projectRoute(project.id);
  return {
    key: `project:${project.id}`,
    label: project.name,
    group: 'projects',
    hint: ['Open', project.name],
    letter: {
      text: project.name.charAt(0).toUpperCase(),
      class: 'bg-background text-foreground ring-1 ring-inset ring-border',
    },
    to,
    enter: { project: entry(project) },
    run: () => router.push(to),
  };
}

function tabItems(resource: PaletteResource): PaletteItem[] {
  return TABS[resource.type].map(tab => ({
    key: `tab:${resource.id}:${tab.key}`,
    label: tab.label,
    group: 'tabs',
    hint: ['Open', tab.label],
    icon: tab.icon,
    run: () => openResource(resource, tab.key),
  }));
}

function serviceUrl(resource: PaletteResource) {
  const endpoints = activeEnvServices.value.find(s => s.id === resource.id)?.endpoints ?? [];
  const reachable = endpoints.filter(e => !e.redirectTo);
  return (reachable.find(e => e.type === EndpointType.Custom) ?? reachable.find(e => e.type === EndpointType.Platform))?.host;
}

function actionItems(resource: PaletteResource): PaletteItem[] {
  if (resource.type !== 'service') return [];
  const items: PaletteItem[] = [];

  if (resource.kind === 'repositoryService') {
    items.push({
      key: `deploy:${resource.id}`,
      label: 'Deploy',
      group: 'actions',
      hint: ['Deploy', resource.name],
      icon: Rocket,
      meta: activeEnvServices.value.find(s => s.id === resource.id)?.branch ?? undefined,
      run: () => {
        deployTarget.value = resource;
        return deploy.startDeploy();
      },
    });
  }

  const host = serviceUrl(resource);
  if (host) {
    items.push({
      key: `url:${resource.id}`,
      label: `Open ${host}`,
      group: 'actions',
      hint: ['Open', host],
      icon: ArrowUpRight,
      run: () => window.open(`https://${host}`, '_blank', 'noopener'),
    });
  }

  return items;
}

function environmentPageItems(environment: PaletteEnvironment): PaletteItem[] {
  const variables = environmentLocation(environment, 'variables');
  const settings = environmentLocation(environment, 'settings');
  return [
    { key: `variables:${environment.id}`, label: 'Shared variables', group: 'pages', hint: ['Open', 'Shared variables'], icon: Braces, to: variables, run: () => router.push(variables) },
    { key: `settings:${environment.id}`, label: 'Environment settings', group: 'pages', hint: ['Open', 'Environment settings'], icon: SlidersHorizontal, to: settings, run: () => router.push(settings) },
  ];
}

function newResourceItem(environment: PaletteEnvironment): PaletteItem {
  return {
    key: `new:${environment.id}`,
    label: 'New',
    group: 'new',
    hint: ['New'],
    icon: Plus,
    keywords: 'create add service database bucket volume',
    shortcut: routeEnvironmentId(route) === environment.id ? 'N' : undefined,
    run: () => router.push({ ...environmentLocation(environment), query: { new: 'resource' } }),
  };
}

function projectPageItems(project: PaletteProject): PaletteItem[] {
  const to = projectRoute(project.id, 'settings');
  return [
    { key: `project-settings:${project.id}`, label: 'Project settings', group: 'pages', hint: ['Open', 'Project settings'], icon: SlidersHorizontal, to, run: () => router.push(to) },
  ];
}

const workspaceItems = computed<PaletteItem[]>(() => {
  const items: PaletteItem[] = workspaceNavItems(isWorkspaceAdmin.value, agents.value).map(item => {
    const label = item.key === 'settings' ? 'Workspace settings' : item.label;
    return { key: `nav:${item.key}`, label, group: 'pages', hint: ['Open', label], icon: item.icon, to: item.to, run: () => router.push(item.to) };
  });

  if (workspaces.value.length > 1) {
    items.push(...workspaces.value.map(ws => ({
      key: `workspace:${ws.id}`,
      label: ws.name,
      group: 'workspaces',
      hint: ['Switch to', ws.name] as [string, string],
      letter: { text: ws.name.charAt(0).toUpperCase(), class: '', style: workspaceColor(ws.id) },
      keywords: 'switch workspace',
      current: ws.id === activeWorkspace.value,
      run: () => handleWorkspaceSwitch(ws.id),
    })));
  }

  const themeLabel = theme.value === 'dark' ? 'Switch to light theme' : 'Switch to dark theme';
  items.push(
    { key: 'theme', label: themeLabel, group: 'app', hint: [themeLabel], icon: theme.value === 'dark' ? Sun : Moon, keywords: 'theme appearance dark light', run: toggleTheme },
    { key: 'docs', label: 'Docs', group: 'app', hint: ['Open', 'Docs'], icon: BookOpen, keywords: 'documentation help', run: () => window.open(DOCS_URL, '_blank', 'noopener') },
    {
      key: 'sign-out',
      label: 'Sign out',
      group: 'app',
      hint: ['Sign out'],
      icon: LogOut,
      keywords: 'log out logout',
      run: async () => {
        await logout();
        router.push('/login');
      },
    },
    {
      key: 'new-project',
      label: 'New project',
      group: 'new',
      hint: ['New project'],
      icon: Plus,
      keywords: 'create add',
      shortcut: route.name === 'projects' ? 'N' : undefined,
      run: () => router.push({ name: 'projects', query: { new: 'project' } }),
    },
  );

  return items;
});

const browseItems = computed<PaletteItem[]>(() => {
  const resource = scopeResource.value;
  if (scope.value.resource) {
    return resource ? [...tabItems(resource), ...actionItems(resource)] : [];
  }

  const environment = scopeEnvironment.value;
  if (scope.value.environment) {
    if (!environment) return [];
    return [
      ...index.value.resources.filter(r => r.environment.id === environment.id).map(r => resourceItem(r)),
      ...environmentPageItems(environment),
      newResourceItem(environment),
    ];
  }

  const project = scopeProject.value;
  if (scope.value.project) {
    if (!project) return [];
    return [
      ...index.value.environments.filter(e => e.project.id === project.id).map(environmentItem),
      ...projectPageItems(project),
    ];
  }

  const recent = recentIds.value
    .map(id => index.value.resources.find(r => r.id === id))
    .filter((r): r is PaletteResource => !!r)
    .slice(0, 3)
    .map(r => resourceItem(r, 'recent'));

  return [
    ...recent,
    ...index.value.projects.map(projectItem),
    ...workspaceItems.value.filter(item => item.key !== 'nav:projects'),
  ];
});

function sameServiceElsewhere(resource: PaletteResource): PaletteItem[] {
  return index.value.resources
    .filter(r => r.environment.project.id === resource.environment.project.id
      && r.environment.id !== resource.environment.id
      && r.type === resource.type
      && r.name === resource.name)
    .map(r => ({
      ...resourceItem(r),
      key: `elsewhere:${r.id}`,
      group: 'elsewhere',
      label: `${r.name} in ${r.environment.name}`,
      matchLabel: r.environment.name,
      keywords: r.name,
      boost: 1,
      meta: undefined,
    }));
}

const searchCandidates = computed<PaletteItem[]>(() => {
  const resource = scopeResource.value;
  const environment = scopeEnvironment.value;
  const project = scopeProject.value;

  const items = [
    ...(resource ? [...sameServiceElsewhere(resource), ...tabItems(resource), ...actionItems(resource)] : []),
    ...index.value.resources.map(r => resourceItem(r)),
    ...index.value.environments.map(environmentItem),
    ...index.value.projects.map(projectItem),
    ...(environment ? [...environmentPageItems(environment), newResourceItem(environment)] : []),
    ...(project ? projectPageItems(project) : []),
    ...workspaceItems.value,
  ];

  const seen = new Set<string>();
  return items.filter(item => !seen.has(item.key) && seen.add(item.key));
});

function proximity(item: PaletteItem) {
  const target = item.enter?.environment?.id;
  if (!target) return 0;
  if (target === scope.value.environment?.id) return 0;
  if (item.enter?.project?.id === scope.value.project?.id) return 1;
  return 2;
}

function labelScore(label: string, term: string) {
  const start = label.toLowerCase().indexOf(term);
  if (start === -1) return 0;
  if (start === 0) return 3;
  return /[\s/_.-]/.test(label.charAt(start - 1)) ? 2 : 1;
}

const terms = computed(() => query.value.trim().toLowerCase().split(/\s+/).filter(Boolean));
const searching = computed(() => terms.value.length > 0);
const helpMode = computed(() => query.value.trim() === '?');

const searchResults = computed<PaletteItem[]>(() => {
  if (!searching.value || helpMode.value) return [];

  return searchCandidates.value
    .map(item => {
      let score = 0;
      for (const term of terms.value) {
        const fromLabel = labelScore(item.matchLabel ?? item.label, term);
        const elsewhere = `${item.keywords ?? ''} ${item.meta ?? ''}`.toLowerCase().includes(term);
        if (!fromLabel && !elsewhere) return null;
        score += fromLabel ? fromLabel + (item.boost ?? 0) : 0.5;
      }
      return { item, score };
    })
    .filter((match): match is { item: PaletteItem; score: number } => match !== null)
    .sort((a, b) =>
      b.score - a.score
      || proximity(a.item) - proximity(b.item)
      || GROUP_ORDER.indexOf(a.item.group) - GROUP_ORDER.indexOf(b.item.group)
      || a.item.label.localeCompare(b.item.label))
    .slice(0, 50)
    .map(match => match.item);
});

const requestItem = computed<PaletteItem>(() => ({
  key: 'request',
  label: 'Request this command',
  group: 'request',
  hint: ['Request this command'],
  icon: MessageSquarePlus,
  run: () => requestFeature({ title: query.value.trim() }),
}));

const noResults = computed(() => searching.value && !helpMode.value && !loading.value && searchResults.value.length === 0);

const items = computed<PaletteItem[]>(() => {
  if (helpMode.value) return [];
  if (noResults.value) return [requestItem.value];
  return searching.value ? searchResults.value : browseItems.value;
});

const focusedItem = computed(() => items.value[focusedIndex.value] ?? null);
const showSkeleton = computed(() => loading.value && !result.value);

function highlight(label: string) {
  const term = terms.value.find(t => label.toLowerCase().includes(t));
  if (!term) return [{ text: label, match: false }];
  const start = label.toLowerCase().indexOf(term);
  return [
    { text: label.slice(0, start), match: false },
    { text: label.slice(start, start + term.length), match: true },
    { text: label.slice(start + term.length), match: false },
  ];
}

function close() {
  open.value = false;
}

function focusInput() {
  nextTick(() => inputRef.value?.focus());
}

function enter(item: PaletteItem) {
  if (!item.enter) return;
  scope.value = item.enter;
  query.value = '';
  focusInput();
}

function up() {
  const { project, environment, resource } = scope.value;
  if (resource) scope.value = { project, environment };
  else if (environment) scope.value = { project };
  else scope.value = {};
}

async function run(item: PaletteItem, newTab = false) {
  if (newTab && item.to) {
    window.open(router.resolve(item.to).href, '_blank', 'noopener');
    return;
  }
  close();
  await item.run();
}

function move(delta: number) {
  const count = items.value.length;
  if (count === 0) return;
  focusedIndex.value = (focusedIndex.value + delta + count) % count;
  nextTick(() => listRef.value?.querySelector('[data-focused="true"]')?.scrollIntoView({ block: 'nearest' }));
}

function onKeydown(event: KeyboardEvent) {
  if (event.isComposing) return;
  const item = focusedItem.value;
  const input = event.target as HTMLInputElement;

  switch (event.key) {
    case 'ArrowDown':
      move(1);
      break;
    case 'ArrowUp':
      move(-1);
      break;
    case 'Enter':
      if (item) run(item, event.metaKey || event.ctrlKey);
      break;
    case 'Tab':
      if (item?.enter) enter(item);
      break;
    case 'ArrowRight':
      if (!item?.enter || input.selectionStart !== input.value.length) return;
      enter(item);
      break;
    case 'Backspace':
      if (query.value || !path.value.length) return;
      up();
      break;
    case 'Escape':
      close();
      break;
    default:
      return;
  }

  event.preventDefault();
  event.stopPropagation();
}

function newFromHere() {
  if (route.name === 'environment') router.push({ query: { ...route.query, new: 'resource' } });
  else if (route.name === 'projects') router.push({ query: { ...route.query, new: 'project' } });
}

useEventListener(window, 'keydown', (event: KeyboardEvent) => {
  if (event.key.toLowerCase() === 'k' && (event.metaKey || event.ctrlKey) && !event.altKey && !event.shiftKey) {
    if (!open.value && document.querySelector(OPEN_LAYER)) return;
    event.preventDefault();
    open.value = !open.value;
    return;
  }

  if (event.key !== 'n' || event.metaKey || event.ctrlKey || event.altKey || open.value) return;
  if (isEditable(event.target) || document.querySelector(OPEN_LAYER)) return;
  event.preventDefault();
  newFromHere();
});

watch(open, (value) => {
  if (!value) return;
  scope.value = scopeFromRoute();
  query.value = '';
  recentIds.value = recentResources(activeWorkspace.value);
  focusInput();
});

watch([query, scope], () => {
  focusedIndex.value = 0;
});

watch(() => route.fullPath, close);

watch(currentPanel, (panel) => {
  if (panel && panel.type !== 'agent' && activeWorkspace.value) rememberResource(activeWorkspace.value, panel.id);
});

const kbdClass = 'rounded border bg-card px-1.5 font-sans text-xs leading-4.5 text-muted-foreground';
</script>

<template>
  <Teleport to="body">
    <Transition name="palette">
      <div
        v-if="open"
        role="dialog"
        aria-modal="true"
        aria-label="Command palette"
        class="fixed inset-0 z-50 flex items-start justify-center pt-[20vh] max-md:px-2 max-md:pt-2"
      >
        <div class="absolute inset-0 bg-background/80 backdrop-blur-sm" @click="close" />

        <div class="relative z-10 w-full max-w-xl overflow-hidden rounded-xl border bg-popover shadow-2xl">
          <div class="flex items-center gap-2 border-b px-3">
            <Search :size="18" class="shrink-0 text-muted-foreground" />
            <div v-if="path.length" class="flex min-w-0 shrink items-center gap-1.5 text-sm">
              <template v-for="segment in path" :key="segment.id">
                <span class="flex min-w-0 items-center gap-1.5">
                  <ResourceIcon v-if="segment.kind" :kind="segment.kind" />
                  <span class="max-w-32 truncate">{{ segment.name }}</span>
                </span>
                <span class="text-muted-foreground">/</span>
              </template>
            </div>
            <input
              ref="inputRef"
              v-model="query"
              placeholder="Search…"
              aria-label="Search"
              class="flex h-12 min-w-0 flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
              @keydown="onKeydown"
            />
            <kbd :class="[kbdClass, 'max-md:hidden']">esc</kbd>
          </div>

          <div ref="listRef" class="max-h-112 overflow-y-auto p-1">
            <template v-if="helpMode">
              <div v-for="row in helpRows" :key="row.label" class="flex items-center justify-between px-2 py-2 text-sm">
                {{ row.label }}
                <kbd :class="kbdClass">{{ row.key }}</kbd>
              </div>
            </template>

            <div v-else-if="showSkeleton" class="space-y-1">
              <Skeleton v-for="i in 4" :key="i" class="h-9 w-full" />
            </div>

            <template v-else>
              <p v-if="noResults" class="px-2 pt-6 pb-4 text-center text-sm text-muted-foreground">
                No results for "{{ query.trim() }}".
              </p>
              <template v-for="(item, i) in items" :key="item.key">
                <div v-if="i > 0 && !searching && item.group !== items[i - 1]!.group" class="-mx-1 my-1 h-px bg-border" />
                <div
                  :data-focused="focusedIndex === i"
                  class="flex items-center rounded-lg pr-1 transition-colors"
                  :class="focusedIndex === i ? 'bg-accent' : 'hover:bg-accent'"
                  @mouseenter="focusedIndex = i"
                >
                  <button
                    type="button"
                    class="flex min-w-0 flex-1 items-center gap-2.5 px-2 py-2.5 text-left text-sm text-popover-foreground"
                    @click="run(item)"
                  >
                    <span class="flex size-4 shrink-0 items-center justify-center">
                      <ResourceIcon v-if="item.resourceKind" :kind="item.resourceKind" :size="16" />
                      <span
                        v-else-if="item.letter"
                        class="flex size-4 items-center justify-center rounded text-xs font-semibold"
                        :class="item.letter.class"
                        :style="item.letter.style"
                      >{{ item.letter.text }}</span>
                      <component :is="item.icon" v-else-if="item.icon" :size="16" class="text-muted-foreground" />
                    </span>
                    <span class="min-w-0 flex-1 truncate">
                      <span
                        v-for="(part, p) in highlight(item.label)"
                        :key="p"
                        :class="part.match ? 'font-semibold' : ''"
                      >{{ part.text }}</span>
                    </span>
                    <span v-if="item.meta" class="shrink-0 text-xs text-muted-foreground">{{ item.meta }}</span>
                    <kbd v-if="item.shortcut" :class="kbdClass">{{ item.shortcut }}</kbd>
                    <Check v-if="item.current" :size="16" class="shrink-0 text-muted-foreground" />
                  </button>
                  <IconButton
                    v-if="item.enter"
                    :label="`Go into ${item.label}`"
                    size="icon-xs"
                    tabindex="-1"
                    @click="enter(item)"
                  >
                    <ChevronRight />
                  </IconButton>
                </div>
              </template>
            </template>
          </div>

          <div v-if="!helpMode" class="flex h-9 items-center gap-4 border-t px-3 text-xs text-muted-foreground max-md:hidden">
            <span v-if="focusedItem" class="flex min-w-0 items-center gap-1.5">
              <kbd :class="kbdClass">↵</kbd>
              <span class="truncate">{{ focusedItem.hint[0] }} <span class="text-foreground">{{ focusedItem.hint[1] }}</span></span>
            </span>
            <span v-if="focusedItem?.enter" class="flex min-w-0 items-center gap-1.5">
              <kbd :class="kbdClass">→</kbd>
              <span class="truncate">Go into <span class="text-foreground">{{ focusedItem.label }}</span></span>
            </span>
            <span class="flex-1" />
            <span v-if="upHint" class="flex shrink-0 items-center gap-1.5">
              <kbd :class="kbdClass">⌫</kbd>
              {{ upHint }}
            </span>
            <span v-else-if="focusedItem?.to" class="flex shrink-0 items-center gap-1.5">
              <kbd :class="kbdClass">{{ modKey }}↵</kbd>
              New tab
            </span>
            <span v-else class="flex shrink-0 items-center gap-1.5">
              <kbd :class="kbdClass">esc</kbd>
              Close
            </span>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>

  <WorkspaceSwitchDialog :workspace-switch="workspaceSwitch" @finished="finishWorkspaceSwitch" />
</template>

<style scoped>
.palette-enter-active,
.palette-leave-active {
  transition: opacity 0.15s ease;
}

.palette-enter-active .relative,
.palette-leave-active .relative {
  transition: transform 0.15s ease, opacity 0.15s ease;
}

.palette-enter-from,
.palette-leave-to {
  opacity: 0;
}

.palette-enter-from .relative,
.palette-leave-to .relative {
  transform: scale(0.96);
  opacity: 0;
}
</style>
