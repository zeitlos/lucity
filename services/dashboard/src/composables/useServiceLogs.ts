import { computed, onScopeDispose, ref, shallowRef, watch, type Ref } from 'vue';
import { refDebounced } from '@vueuse/core';
import { useApolloClient, useQuery, useSubscription } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { LogDirection, type ServiceLogFilter, type ServiceLogsQuery } from '@/gql/graphql';
import { LIVE_WINDOW, instanceColors, rangeSpans, type LogRange } from '@/lib/logs';

const ServiceLogsDocument = graphql(`
  query ServiceLogs($service: ServiceID!, $from: Time!, $to: Time, $filter: ServiceLogFilter, $surrounding: Int, $limit: Int, $direction: LogDirection) {
    serviceLogs(service: $service, from: $from, to: $to, filter: $filter, surrounding: $surrounding, limit: $limit, direction: $direction) {
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

export type LogChange = 'reset' | 'earlier' | 'later';

function timeOf(entry: LogEntry): number {
  return Date.parse(entry.time);
}

function iso(at: number): string {
  return new Date(at).toISOString();
}

const PAGE = 1000;
const HALF_PAGE = 500;
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

  const timeline = useQuery(
    ServiceLogTimelineDocument,
    () => ({
      service: serviceId.value,
      from: iso(bounds.value.from),
      to: iso(bounds.value.to),
      filter: filter.value,
    }),
    () => ({ enabled: !!serviceId.value, fetchPolicy: 'no-cache' }),
  );

  const query = computed(() => ({
    service: serviceId.value,
    filter: filter.value,
    surrounding: searching.value && surrounding.value ? SURROUNDING : 0,
  }));

  const historyEntries = shallowRef<LogEntry[]>([]);
  const historyLoading = ref(false);
  const historyError = ref<unknown>(null);
  const moreBefore = ref(false);
  const moreAfter = ref(false);
  const loadingEarlier = ref(false);
  const loadingLater = ref(false);
  const change = ref<LogChange>('reset');
  const target = ref<number | null>(null);
  const version = ref(0);
  const { resolveClient } = useApolloClient();
  let request = 0;
  let pendingSeek: number | null = null;

  async function fetchPage(from: number, to: number, limit: number, direction: LogDirection): Promise<LogEntry[]> {
    const { data } = await resolveClient().query({
      query: ServiceLogsDocument,
      variables: { ...query.value, from: iso(from), to: iso(to), limit, direction },
      fetchPolicy: 'no-cache',
    });
    return data?.serviceLogs ?? [];
  }

  function apply(next: LogEntry[], kind: LogChange, at: number | null) {
    historyEntries.value = next;
    change.value = kind;
    target.value = at;
    version.value++;
  }

  async function loadAround(at: number | null) {
    const id = ++request;
    const { from, to } = bounds.value;
    historyLoading.value = true;
    historyError.value = null;
    try {
      if (at === null) {
        const page = await fetchPage(from, to, PAGE, LogDirection.Backward);
        if (id !== request) return;
        moreBefore.value = page.length >= PAGE;
        moreAfter.value = false;
        apply(page, 'reset', null);
      } else {
        const [before, after] = await Promise.all([
          fetchPage(from, at, HALF_PAGE, LogDirection.Backward),
          fetchPage(at, to, HALF_PAGE, LogDirection.Forward),
        ]);
        if (id !== request) return;
        moreBefore.value = before.length >= HALF_PAGE;
        moreAfter.value = after.length >= HALF_PAGE;
        apply([...before, ...after.filter(entry => timeOf(entry) > at)], 'reset', at);
      }
    } catch (cause) {
      if (id === request) historyError.value = cause;
    } finally {
      if (id === request) historyLoading.value = false;
    }
  }

  watch([query, bounds, live], () => {
    if (live.value || !serviceId.value) return;
    const at = pendingSeek;
    pendingSeek = null;
    void loadAround(at);
  }, { immediate: true });

  async function loadEarlier() {
    const first = historyEntries.value[0];
    if (!first || loadingEarlier.value) return;
    const id = request;
    loadingEarlier.value = true;
    try {
      const page = await fetchPage(bounds.value.from, timeOf(first), PAGE, LogDirection.Backward);
      if (id !== request) return;
      moreBefore.value = page.length >= PAGE;
      apply([...page.filter(entry => timeOf(entry) < timeOf(first)), ...historyEntries.value], 'earlier', null);
    } finally {
      loadingEarlier.value = false;
    }
  }

  async function loadLater() {
    const last = historyEntries.value[historyEntries.value.length - 1];
    if (!last || loadingLater.value) return;
    const id = request;
    loadingLater.value = true;
    try {
      const page = await fetchPage(timeOf(last), bounds.value.to, PAGE, LogDirection.Forward);
      if (id !== request) return;
      moreAfter.value = page.length >= PAGE;
      apply([...historyEntries.value, ...page.filter(entry => timeOf(entry) > timeOf(last))], 'later', null);
    } finally {
      loadingLater.value = false;
    }
  }

  function seek(at: number) {
    if (live.value) {
      pendingSeek = at;
      historyEntries.value = [];
      setRange('custom', { ...bounds.value });
      return;
    }
    void loadAround(at);
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

  const loading = computed(() => (live.value ? !liveStarted.value && !subscription.error.value : historyLoading.value && !historyEntries.value.length));
  const error = computed(() => (live.value ? subscription.error.value : historyError.value) ?? null);

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
    else void loadAround(null);
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
    hasEarlier: computed(() => !live.value && moreBefore.value),
    hasLater: computed(() => !live.value && moreAfter.value),
    loadingEarlier,
    loadingLater,
    change,
    target,
    version,
    setRange,
    loadEarlier,
    loadLater,
    seek,
    colorOf,
    retry,
  };
}
