<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { ArrowDown, ArrowUp, Ban, Copy, RefreshCw, Search, X } from '@lucide/vue';

import FilterBar from '@/components/FilterBar.vue';
import MarkAbuseDialog from '@/components/MarkAbuseDialog.vue';
import ResultToast from '@/components/ResultToast.vue';
import SignalIcons from '@/components/SignalIcons.vue';
import TrendCharts from '@/components/TrendCharts.vue';
import WorkspacePeek from '@/components/WorkspacePeek.vue';
import { Plan } from '@/gql/graphql';
import { age, gigabytes, loadedTime, memory, relativeTime } from '@/lib/format';
import {
  CachedWorkspacesDocument,
  ClearAbuseDocument,
  MarkAbuseDocument,
  WorkspacesDocument,
  avatarStyle,
  githubProfile,
  hasStrongSignal,
  initial,
  matches,
  openExternal,
  owner,
  plan,
  status,
  toneClass,
  youngAccount,
  type Filters,
  type Outcome,
  type Workspace,
} from '@/lib/workspaces';

const columns = 'grid-cols-[32px_minmax(160px,1.4fr)_150px_96px_170px_minmax(190px,1.3fr)_80px_64px_84px_150px]';

const lowEgress = 5e9;
const lowMemory = 2 ** 30;

function trial(workspace: Workspace): boolean {
  return workspace.billing?.plan === Plan.None && workspace.billing.goodStanding;
}

function paid(workspace: Workspace): boolean {
  return workspace.billing != null && workspace.billing.plan !== Plan.None;
}

const polling = ref(false);

const { result: cached } = useQuery(CachedWorkspacesDocument);
const { result, loading, error, refetch } = useQuery(WorkspacesDocument, null, () => ({
  pollInterval: polling.value ? 60_000 : 0,
  notifyOnNetworkStatusChange: true,
}));

const list = computed(() => result.value?.workspaces ?? cached.value?.cachedWorkspaces ?? null);
const workspaces = computed(() => list.value?.workspaces ?? []);

const loadState = computed(() => {
  const loaded = list.value ? `Loaded ${loadedTime(new Date(list.value.loadedAt))}` : '';

  if (loading.value) return loaded ? `${loaded}, refreshing…` : 'Loading…';
  if (error.value && list.value) return `${loaded}, refresh failed`;
  return loaded;
});

watch(workspaces, (list) => {
  polling.value = list.some((workspace) => workspace.abuse && !workspace.suspended);
});

const filters = ref<Filters>({ choices: {}, ranges: {} });
const search = ref('');
const searchInput = useTemplateRef<HTMLInputElement>('searchInput');

function haystack(workspace: Workspace): string {
  const parts = [workspace.id];

  for (const member of workspace.members) {
    parts.push(member.name ?? '', member.email ?? '', member.account?.login ?? '');
  }

  for (const environment of workspace.environments) {
    for (const service of environment.services) {
      parts.push(service.image ?? '', service.commit ?? '', service.repository ?? '');
    }
  }

  return parts.join('').toLowerCase();
}

type SortColumn = {
  key: string;
  label: string;
  numeric?: boolean;
  descending: boolean;
  value: (workspace: Workspace) => number | string;
};

const planRank = { [Plan.None]: 1, [Plan.Hobby]: 2, [Plan.Pro]: 3 };

const sortColumns: SortColumn[] = [
  { key: 'workspace', label: 'Workspace', descending: false, value: (workspace) => workspace.id },
  { key: 'created', label: 'Created', descending: true, value: (workspace) => Date.parse(workspace.createdAt) },
  {
    key: 'signals',
    label: 'Signals',
    descending: true,
    value: (workspace) => workspace.signals.reduce((sum, signal) => sum + (signal.strong ? 10 : 1), 0),
  },
  { key: 'plan', label: 'Plan', descending: true, value: (workspace) => (workspace.billing ? planRank[workspace.billing.plan] : 0) },
  {
    key: 'owner',
    label: 'Owner',
    descending: false,
    value: (workspace) => (owner(workspace)?.account?.login ?? owner(workspace)?.email ?? '').toLowerCase(),
  },
  { key: 'egress', label: 'Egress 7d', numeric: true, descending: true, value: (workspace) => workspace.egress7d },
  { key: 'pods', label: 'Pods', numeric: true, descending: true, value: (workspace) => workspace.runningPods + workspace.pendingPods },
  { key: 'requests', label: 'Requests', numeric: true, descending: true, value: (workspace) => workspace.memoryRequested },
  { key: 'status', label: 'Status', descending: true, value: (workspace) => (workspace.suspended ? 2 : workspace.abuse ? 1 : 0) },
];

const sort = ref({ key: 'created', descending: true });

function sortBy(column: SortColumn) {
  sort.value = sort.value.key === column.key
    ? { key: column.key, descending: !sort.value.descending }
    : { key: column.key, descending: column.descending };
}

function compare(left: number | string, right: number | string): number {
  if (typeof left === 'number' && typeof right === 'number') return left - right;
  return String(left).localeCompare(String(right));
}

const visible = computed(() => {
  const term = search.value.trim().toLowerCase();
  const column = sortColumns.find((candidate) => candidate.key === sort.value.key)!;
  const direction = sort.value.descending ? -1 : 1;

  return workspaces.value
    .filter((workspace) => matches(workspace, filters.value) && (!term || haystack(workspace).includes(term)))
    .sort((left, right) => compare(column.value(left), column.value(right)) * direction);
});

const selected = ref(new Set<string>());
const selectedWorkspaces = computed(() => workspaces.value.filter((workspace) => selected.value.has(workspace.id)));
const allVisibleSelected = computed(() => visible.value.length > 0 && visible.value.every((workspace) => selected.value.has(workspace.id)));
const someVisibleSelected = computed(() => !allVisibleSelected.value && visible.value.some((workspace) => selected.value.has(workspace.id)));
const withSignals = computed(() => workspaces.value.filter((workspace) => hasStrongSignal(workspace) && !workspace.abuse));
const allWithSignalsSelected = computed(() => withSignals.value.length > 0 && withSignals.value.every((workspace) => selected.value.has(workspace.id)));

function toggle(id: string) {
  const next = new Set(selected.value);

  if (next.has(id)) next.delete(id);
  else next.add(id);

  selected.value = next;
}

function toggleVisible() {
  const next = new Set(selected.value);

  for (const workspace of visible.value) {
    if (allVisibleSelected.value) next.delete(workspace.id);
    else next.add(workspace.id);
  }

  selected.value = next;
}

function selectWithSignals() {
  selected.value = new Set([...selected.value, ...withSignals.value.map((workspace) => workspace.id)]);
}

function clearSelection() {
  selected.value = new Set();
}

const selectionPods = computed(() => selectedWorkspaces.value.reduce((sum, workspace) => sum + workspace.runningPods + workspace.pendingPods, 0));
const selectionMemory = computed(() => selectedWorkspaces.value.reduce((sum, workspace) => sum + workspace.memoryRequested, 0));

async function copyNames() {
  await navigator.clipboard.writeText(selectedWorkspaces.value.map((workspace) => workspace.id).join('\n'));
}

const peekID = ref<string | null>(null);
const peekIndex = computed(() => visible.value.findIndex((workspace) => workspace.id === peekID.value));
const peekWorkspace = computed(() => workspaces.value.find((workspace) => workspace.id === peekID.value) ?? null);

function step(offset: number) {
  if (visible.value.length === 0) return;

  const index = Math.min(Math.max(peekIndex.value + offset, 0), visible.value.length - 1);
  peekID.value = visible.value[index]!.id;
}

const dialogWorkspaces = ref<Workspace[] | null>(null);
const outcome = ref<Outcome | null>(null);
const lastNote = ref('');

const { mutate: markAbuse, loading: marking } = useMutation(MarkAbuseDocument);
const { mutate: clearAbuse, loading: clearing } = useMutation(ClearAbuseDocument);

function failedAll(workspaceIDs: string[], e: unknown): Outcome['results'] {
  const message = e instanceof Error ? e.message : String(e);
  return workspaceIDs.map((workspaceID) => ({ workspaceID, error: message }));
}

async function mark(workspaceIDs: string[], note: string) {
  lastNote.value = note;

  try {
    const response = await markAbuse({ workspaceIDs, note: note || null });
    outcome.value = { verb: 'marked', results: response?.data?.markAbuse ?? [] };
  } catch (e) {
    outcome.value = { verb: 'marked', results: failedAll(workspaceIDs, e) };
  }

  dialogWorkspaces.value = null;
  clearSelection();
  await refetch();
}

async function clear(workspaceIDs: string[]) {
  try {
    const response = await clearAbuse({ workspaceIDs });
    outcome.value = { verb: 'cleared', results: response?.data?.clearAbuse ?? [] };
  } catch (e) {
    outcome.value = { verb: 'cleared', results: failedAll(workspaceIDs, e) };
  }

  await refetch();
}

function retry(workspaceIDs: string[]) {
  if (outcome.value?.verb === 'marked') mark(workspaceIDs, lastNote.value);
  else clear(workspaceIDs);
}

function onKeydown(event: KeyboardEvent) {
  const target = event.target as HTMLElement | null;

  if (dialogWorkspaces.value || target?.closest('input, textarea') || document.querySelector(':popover-open')) return;

  if (event.key === '/') {
    event.preventDefault();
    searchInput.value?.focus();
  } else if (peekID.value && event.key === 'j') {
    step(1);
  } else if (peekID.value && event.key === 'k') {
    step(-1);
  } else if (event.key === 'Escape') {
    peekID.value = null;
  }
}

const filterBar = useTemplateRef<HTMLElement>('filterBar');
const filterBarHeight = ref(48);
const filterBarObserver = new ResizeObserver(([entry]) => {
  filterBarHeight.value = entry!.borderBoxSize[0]!.blockSize;
});

onMounted(() => {
  window.addEventListener('keydown', onKeydown);
  filterBarObserver.observe(filterBar.value!);
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown);
  filterBarObserver.disconnect();
});
</script>

<template>
  <main class="relative flex min-h-0 flex-1 flex-col bg-card">
    <div class="relative flex min-h-0 flex-1 flex-col">
      <div class="min-h-0 flex-1 overflow-auto">
        <TrendCharts :loaded-at="list?.loadedAt" />

        <div ref="filterBar" class="sticky top-0 left-0 z-20 flex min-h-12 flex-wrap items-center gap-3 bg-card px-4 py-2">
          <FilterBar v-model="filters" :workspaces="workspaces" />
          <div class="flex-1" />
          <label class="flex h-8 min-w-[280px] items-center gap-2 rounded-sm border bg-background px-2.5 text-muted-foreground">
            <Search class="size-3.5" aria-hidden="true" />
            <span class="sr-only">Filter workspaces</span>
            <input
              ref="searchInput"
              v-model="search"
              type="search"
              placeholder="Search workspaces"
              class="min-w-0 flex-1 bg-transparent text-foreground outline-none"
            >
            <kbd class="rounded-xs border px-1.5 font-mono">/</kbd>
          </label>
        </div>

        <div v-if="error && !list" class="sticky left-0 m-4 rounded-sm border border-status-danger/40 px-4 py-3">
          <p class="font-medium text-status-danger">Workspaces could not be loaded</p>
          <p class="mt-1 text-muted-foreground">{{ error.message }}</p>
        </div>

        <div v-else class="min-w-[1240px]">
          <div
            class="sticky z-10 grid h-9 items-center gap-x-3 border-y bg-sidebar px-4 text-muted-foreground"
            :class="columns"
            :style="{ top: `${filterBarHeight}px` }"
          >
            <input
              type="checkbox"
              class="size-4 accent-foreground"
              :checked="allVisibleSelected"
              :indeterminate="someVisibleSelected"
              aria-label="Select all shown"
              @change="toggleVisible"
            >
            <button
              v-for="column in sortColumns"
              :key="column.key"
              type="button"
              class="flex h-full min-w-0 items-center gap-1 hover:text-foreground"
              :class="[column.numeric ? 'flex-row-reverse' : '', sort.key === column.key ? 'font-medium text-foreground' : '']"
              @click="sortBy(column)"
            >
              <span class="truncate">{{ column.label }}</span>
              <component
                :is="sort.descending ? ArrowDown : ArrowUp"
                v-if="sort.key === column.key"
                class="size-3.5 shrink-0"
                aria-hidden="true"
              />
            </button>
          </div>

          <template v-if="!list">
            <div v-for="index in 16" :key="index" class="grid min-h-10 items-center gap-x-3 border-b border-hairline px-4" :class="columns">
              <span />
              <span class="h-4 w-32 animate-pulse rounded-md bg-muted" />
              <span class="h-4 w-40 animate-pulse rounded-md bg-muted" />
              <span class="h-4 w-24 animate-pulse rounded-md bg-muted" />
              <span class="h-4 w-28 animate-pulse rounded-md bg-muted" />
            </div>
          </template>

          <div
            v-for="workspace in visible"
            :key="workspace.id"
            class="grid min-h-10 cursor-pointer items-center gap-x-3 border-b border-hairline px-4 py-1.5"
            :class="[ columns,
              selected.has(workspace.id)
                ? 'bg-selected shadow-[inset_3px_0_0_var(--foreground)]'
                : peekID === workspace.id
                  ? 'bg-accent'
                  : workspace.suspended
                    ? 'bg-muted/70'
                    : 'hover:bg-background/60',
              workspace.suspended ? 'text-muted-foreground' : '',
            ]"
            @click="peekID = workspace.id"
          >
            <input
              type="checkbox"
              class="size-4 accent-foreground"
              :checked="selected.has(workspace.id)"
              :aria-label="`Select ${workspace.id}`"
              @click.stop
              @change="toggle(workspace.id)"
            >
            <div class="flex min-w-0 items-center gap-2.5">
              <span class="flex size-6 shrink-0 items-center justify-center rounded-xs font-semibold" :style="avatarStyle(workspace.id)" aria-hidden="true">
                {{ initial(workspace.id) }}
              </span>
              <span class="truncate font-medium">{{ workspace.id }}</span>
            </div>
            <span class="truncate text-muted-foreground">{{ relativeTime(workspace.createdAt) }}</span>
            <SignalIcons :signals="workspace.signals" :muted="workspace.suspended" />
            <span
              class="truncate"
              :class="[trial(workspace) ? 'text-muted-foreground' : toneClass[plan(workspace).tone], paid(workspace) ? 'font-medium' : '']"
            >{{ plan(workspace).text }}</span>
            <span class="flex min-w-0 items-baseline gap-1.5">
              <template v-if="owner(workspace)?.account">
                <a
                  :href="githubProfile(owner(workspace)!.account!.login)"
                  class="truncate hover:underline"
                  :class="youngAccount(workspace) ? '' : 'text-muted-foreground'"
                  @click.prevent.stop="openExternal(githubProfile(owner(workspace)!.account!.login))"
                >{{ owner(workspace)!.account!.login }}</a>
                <span class="shrink-0" :class="youngAccount(workspace) && !workspace.suspended ? 'text-status-warn' : 'text-muted-foreground'">
                  {{ age(owner(workspace)!.account!.createdAt) }}
                </span>
              </template>
              <span v-else class="truncate text-muted-foreground">{{ owner(workspace)?.email ?? 'No members' }}</span>
              <span v-if="workspace.members.length > 1" class="shrink-0 text-muted-foreground">+{{ workspace.members.length - 1 }}</span>
            </span>
            <span class="text-right tabular-nums" :class="workspace.egress7d < lowEgress ? 'text-muted-foreground' : 'font-medium'">{{ gigabytes(workspace.egress7d) }}</span>
            <span class="text-right tabular-nums">
              {{ workspace.runningPods }}<span v-if="workspace.pendingPods" class="text-status-warn"> +{{ workspace.pendingPods }}</span>
            </span>
            <span class="text-right tabular-nums" :class="workspace.memoryRequested < lowMemory ? 'text-muted-foreground' : 'font-medium'">{{ memory(workspace.memoryRequested) }}</span>
            <span class="truncate">
              <span :class="toneClass[status(workspace).tone]">{{ status(workspace).text }}</span>
              <span v-if="status(workspace).note" class="text-muted-foreground"> · {{ status(workspace).note }}</span>
            </span>
          </div>

          <p v-if="list && visible.length === 0" class="px-4 py-10 text-center text-muted-foreground">No workspaces match.</p>
          <div class="h-24" />
        </div>
      </div>

      <div
        v-if="selected.size > 0"
        class="pointer-events-none absolute bottom-6 left-0 z-30 flex justify-center"
        :class="peekWorkspace ? 'right-[480px]' : 'right-0'"
      >
        <div class="pointer-events-auto flex items-center gap-3 rounded-lg bg-foreground py-2 pr-2 pl-4 whitespace-nowrap text-background shadow-lg">
          <span class="flex flex-col gap-0.5">
            <span class="font-medium">{{ selected.size }} selected</span>
            <span class="opacity-80">{{ selectionPods }} pods · {{ memory(selectionMemory) }} requested</span>
          </span>
          <button v-if="!allWithSignalsSelected && withSignals.length" type="button" class="h-8 rounded-sm px-2.5 underline opacity-90" @click="selectWithSignals">
            Select all {{ withSignals.length }} with signals
          </button>
          <span class="h-6 w-px bg-background/30" aria-hidden="true" />
          <button type="button" class="flex h-8 items-center gap-1.5 rounded-sm border border-background/30 px-3" @click="copyNames">
            <Copy class="size-3.5" aria-hidden="true" />
            Copy names
          </button>
          <button type="button" class="flex h-8 items-center gap-2 rounded-sm bg-destructive px-3 font-medium text-white" @click="dialogWorkspaces = selectedWorkspaces">
            <Ban class="size-3.5" aria-hidden="true" />
            Mark as abuse
          </button>
          <button type="button" class="flex size-8 items-center justify-center rounded-sm opacity-85" aria-label="Clear selection" @click="clearSelection">
            <X class="size-4" aria-hidden="true" />
          </button>
        </div>
      </div>

      <WorkspacePeek
        v-if="peekWorkspace"
        :workspace="peekWorkspace"
        :position="peekIndex + 1"
        :total="visible.length"
        @close="peekID = null"
        @previous="step(-1)"
        @next="step(1)"
        @mark="dialogWorkspaces = [peekWorkspace]"
        @clear="clear([peekWorkspace.id])"
      />

      <MarkAbuseDialog
        v-if="dialogWorkspaces"
        :workspaces="dialogWorkspaces"
        :running="marking"
        @close="dialogWorkspaces = null"
        @confirm="mark"
      />

      <ResultToast
        v-if="outcome"
        :outcome="outcome"
        :running="marking || clearing"
        @dismiss="outcome = null"
        @undo="clear"
        @retry="retry"
      />
    </div>

    <footer class="flex h-7 shrink-0 items-center gap-2 border-t bg-sidebar px-4 text-xs text-muted-foreground">
      <span class="shrink-0 whitespace-nowrap">{{ visible.length }} of {{ workspaces.length }}</span>
      <span
        class="ml-auto min-w-0 truncate"
        :class="error && list ? 'text-status-danger' : ''"
        :title="error && list ? error.message : undefined"
      >{{ loadState }}</span>
      <button
        type="button"
        class="flex size-5 items-center justify-center rounded-md hover:bg-muted hover:text-foreground disabled:opacity-50"
        aria-label="Reload"
        :disabled="loading"
        @click="refetch()"
      >
        <RefreshCw class="size-3" :class="loading ? 'animate-spin' : ''" aria-hidden="true" />
      </button>
    </footer>
  </main>
</template>
