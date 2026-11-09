<script setup lang="ts">
import { computed, onActivated, onDeactivated, ref, watch } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { RefreshCw } from '@lucide/vue';

import PlatformHero from '@/components/PlatformHero.vue';
import PlatformLimits from '@/components/PlatformLimits.vue';
import PlatformLogs from '@/components/PlatformLogs.vue';
import PlatformNodePools from '@/components/PlatformNodePools.vue';
import PlatformRoutines from '@/components/PlatformRoutines.vue';
import { loadedTime } from '@/lib/format';
import {
  AppChartVersionsDocument,
  LimitsDocument,
  NodePoolsDocument,
  PlatformLogsDocument,
  PlatformTotalsDocument,
  RoutinesDocument,
} from '@/lib/platform';

const REFRESH = 60_000;

const active = ref(false);
const options = () => ({ pollInterval: active.value ? REFRESH : 0, notifyOnNetworkStatusChange: true });

const totals = useQuery(PlatformTotalsDocument, null, options);
const versions = useQuery(AppChartVersionsDocument, null, options);
const routines = useQuery(RoutinesDocument, null, options);
const limits = useQuery(LimitsDocument, null, options);
const pools = useQuery(NodePoolsDocument, null, options);
const logs = useQuery(PlatformLogsDocument, null, options);

const queries = [totals, versions, routines, limits, pools, logs];
const loading = computed(() => queries.some((query) => query.loading.value));
const failed = computed(() => queries.filter((query) => query.error.value).length);
const loadedAt = ref<Date | null>(null);

watch(loading, (now) => {
  if (!now) loadedAt.value = new Date();
});

const loadState = computed(() => {
  const loaded = loadedAt.value ? `Loaded ${loadedTime(loadedAt.value)}` : '';
  const problems = failed.value ? `, ${failed.value} ${failed.value === 1 ? 'section' : 'sections'} failed` : '';

  if (loading.value) return loaded ? `${loaded}${problems}, refreshing…` : 'Loading…';
  return `${loaded}${problems}`;
});

function reload() {
  for (const query of queries) query.refetch();
}

onActivated(() => {
  active.value = true;

  if (loadedAt.value && Date.now() - loadedAt.value.getTime() > REFRESH) reload();
});

onDeactivated(() => {
  active.value = false;
});
</script>

<template>
  <main class="flex min-h-0 flex-1 flex-col bg-card">
    <div class="min-h-0 flex-1 overflow-auto">
      <PlatformHero
        :totals="totals.result.value?.platformTotals ?? null"
        :totals-error="totals.error.value?.message ?? null"
        :versions="versions.result.value?.appChartVersions ?? null"
        :versions-error="versions.error.value?.message ?? null"
      />
      <PlatformRoutines :routines="routines.result.value?.routines ?? null" :error="routines.error.value?.message ?? null" />
      <PlatformLimits :groups="limits.result.value?.limits ?? null" :error="limits.error.value?.message ?? null" />
      <PlatformNodePools :pools="pools.result.value?.nodePools ?? null" :error="pools.error.value?.message ?? null" />
      <PlatformLogs :logs="logs.result.value?.platformLogs ?? null" :error="logs.error.value?.message ?? null" />
      <div class="h-28" />
    </div>

    <footer class="flex h-7 shrink-0 items-center gap-2 border-t bg-sidebar px-4 text-xs text-muted-foreground">
      <span class="ml-auto min-w-0 truncate" :class="failed ? 'text-status-danger' : ''">{{ loadState }}</span>
      <button
        type="button"
        class="flex size-5 items-center justify-center rounded-md hover:bg-muted hover:text-foreground disabled:opacity-50"
        aria-label="Reload"
        :disabled="loading"
        @click="reload"
      >
        <RefreshCw class="size-3" :class="loading ? 'animate-spin' : ''" aria-hidden="true" />
      </button>
    </footer>
  </main>
</template>
