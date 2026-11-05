import { computed, onScopeDispose, ref, shallowRef, watch, type Ref } from 'vue';
import { refDebounced } from '@vueuse/core';
import { useApolloClient, useQuery, useSubscription } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import type { ServiceLogFilter, ServiceLogsQuery } from '@/gql/graphql';
import { LIVE_WINDOW, instanceColors, rangeSpans, type LogRange } from '@/lib/logs';

const ServiceLogsDocument = graphql(`
  query ServiceLogs($service: ServiceID!, $from: Time!, $to: Time, $filter: ServiceLogFilter, $surrounding: Int, $limit: Int) {
    serviceLogs(service: $service, from: $from, to: $to, filter: $filter, surrounding: $surrounding, limit: $limit) {
      time
      line
      pod
      stream
      event
    }
  }
`);

const ServiceLogTimelineDocument = graphql(`
  query ServiceLogTimeline($service: ServiceID!, $from: Time!, $to: Time, $filter: ServiceLogFilter) {
    serviceLogInstances(service: $service, from: $from, to: $to) {
      pod
      firstSeen
      lastSeen
      restarts
      state
    }
    serviceLogVolume(service: $service, from: $from, to: $to, filter: $filter) {
      time
      stdout
      stderr
    }
  }
`);

const ServiceLogsLiveDocument = graphql(`
  subscription ServiceLogsLive($service: ServiceID!, $tailLines: Int, $filter: ServiceLogFilter) {
    serviceLogs(service: $service, tailLines: $tailLines, filter: $filter) {
      time
      line
      pod
      stream
      event
    }
  }
`);

export type LogEntry = ServiceLogsQuery['serviceLogs'][number];

const PAGE = 1000;
const LIVE_LIMIT = 5000;
const LIVE_TAIL = 500;
const LIVE_REFRESH = 15_000;
const SURROUNDING = 3;

export function useServiceLogs(serviceId: Ref<string>) {
  const range = ref<LogRange>('live');
  const custom = ref<{ from: number; to: number } | null>(null);
  const search = ref('');
  const pod = ref<string | null>(null);
  const surrounding = ref(false);
  const anchor = ref(Date.now());

  const debouncedSearch = refDebounced(search, 300);
  const live = computed(() => range.value === 'live');

  const bounds = computed(() => {
    if (range.value === 'custom' && custom.value) return custom.value;
    const span = range.value === 'live' || range.value === 'custom' ? LIVE_WINDOW : rangeSpans[range.value];
    return { from: anchor.value - span, to: anchor.value };
  });

  const filter = computed<ServiceLogFilter>(() => ({
    search: debouncedSearch.value.trim() || null,
    pod: pod.value,
  }));

  const searching = computed(() => !!filter.value.search);

  function setRange(next: LogRange, span?: { from: number; to: number }) {
    if (next === 'custom' && span) custom.value = span;
    anchor.value = Date.now();
    range.value = next;
  }

  let refresh: ReturnType<typeof setInterval> | undefined;
  watch(live, isLive => {
    clearInterval(refresh);
    if (isLive) refresh = setInterval(() => { anchor.value = Date.now(); }, LIVE_REFRESH);
  }, { immediate: true });
  onScopeDispose(() => clearInterval(refresh));

  const historyVariables = computed(() => ({
    service: serviceId.value,
    from: new Date(bounds.value.from).toISOString(),
    to: new Date(bounds.value.to).toISOString(),
    filter: filter.value,
    surrounding: searching.value && surrounding.value ? SURROUNDING : 0,
    limit: PAGE,
  }));

  const historyQuery = useQuery(ServiceLogsDocument, historyVariables, () => ({
    enabled: !live.value && !!serviceId.value,
    fetchPolicy: 'no-cache',
  }));

  const timeline = useQuery(
    ServiceLogTimelineDocument,
    () => ({
      service: serviceId.value,
      from: new Date(bounds.value.from).toISOString(),
      to: new Date(bounds.value.to).toISOString(),
      filter: filter.value,
    }),
    () => ({ enabled: !!serviceId.value, fetchPolicy: 'no-cache' }),
  );

  const earlier = shallowRef<LogEntry[]>([]);
  const earlierDone = ref(false);
  const loadingEarlier = ref(false);
  const { resolveClient } = useApolloClient();

  watch(historyVariables, () => {
    earlier.value = [];
    earlierDone.value = false;
  });

  const historyEntries = computed(() => [...earlier.value, ...(historyQuery.result.value?.serviceLogs ?? [])]);

  const hasEarlier = computed(() => {
    if (live.value || earlierDone.value) return false;
    return (historyQuery.result.value?.serviceLogs.length ?? 0) >= PAGE;
  });

  async function loadEarlier() {
    const first = historyEntries.value[0];
    if (!first || loadingEarlier.value) return;
    loadingEarlier.value = true;
    try {
      const { data } = await resolveClient().query({
        query: ServiceLogsDocument,
        variables: { ...historyVariables.value, to: first.time },
        fetchPolicy: 'no-cache',
      });
      const page = (data?.serviceLogs ?? []).filter(entry => entry.time < first.time);
      earlier.value = [...page, ...earlier.value];
      if ((data?.serviceLogs.length ?? 0) < PAGE) earlierDone.value = true;
    } finally {
      loadingEarlier.value = false;
    }
  }

  const liveEntries = shallowRef<LogEntry[]>([]);
  const liveStarted = ref(false);

  const subscription = useSubscription(
    ServiceLogsLiveDocument,
    () => ({ service: serviceId.value, tailLines: LIVE_TAIL, filter: filter.value }),
    () => ({ enabled: live.value && !!serviceId.value }),
  );

  watch([serviceId, filter, live], () => {
    liveEntries.value = [];
    liveStarted.value = false;
  });

  let pending: LogEntry[] = [];
  let flush: number | undefined;
  subscription.onResult(({ data }) => {
    if (!data?.serviceLogs) return;
    liveStarted.value = true;
    pending.push(data.serviceLogs);
    flush ??= requestAnimationFrame(() => {
      const next = [...liveEntries.value, ...pending];
      liveEntries.value = next.length > LIVE_LIMIT ? next.slice(next.length - LIVE_LIMIT) : next;
      pending = [];
      flush = undefined;
    });
  });
  onScopeDispose(() => {
    if (flush !== undefined) cancelAnimationFrame(flush);
  });

  const entries = computed(() => (live.value ? liveEntries.value : historyEntries.value));
  const instances = computed(() => timeline.result.value?.serviceLogInstances ?? []);
  const volume = computed(() => timeline.result.value?.serviceLogVolume ?? []);

  const loading = computed(() => (live.value ? !liveStarted.value && !subscription.error.value : historyQuery.loading.value && !historyQuery.result.value));
  const error = computed(() => (live.value ? subscription.error.value : historyQuery.error.value) ?? null);

  const colors = computed(() => {
    const order = new Map<string, string>();
    const assign = (name: string) => {
      if (!order.has(name)) order.set(name, instanceColors[order.size % instanceColors.length]!);
    };
    instances.value.forEach(instance => assign(instance.pod));
    entries.value.forEach(entry => assign(entry.pod));
    return order;
  });

  function colorOf(name: string): string {
    return colors.value.get(name) ?? 'inherit';
  }

  watch(serviceId, () => {
    search.value = '';
    pod.value = null;
    surrounding.value = false;
    custom.value = null;
    setRange('live');
  });

  function retry() {
    if (live.value) subscription.restart();
    else void historyQuery.refetch();
    void timeline.refetch();
  }

  return {
    range,
    custom,
    search,
    pod,
    surrounding,
    bounds,
    live,
    searching,
    filterSearch: computed(() => filter.value.search ?? ''),
    entries,
    instances,
    volume,
    loading,
    error,
    timelineLoading: computed(() => timeline.loading.value && !timeline.result.value),
    hasEarlier,
    loadingEarlier,
    setRange,
    loadEarlier,
    colorOf,
    retry,
  };
}
