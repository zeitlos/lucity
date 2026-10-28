<script setup lang="ts">
import { serviceStatus } from '@/lib/status';
import { formatDuration } from '@/lib/format';
import { devicon } from '@/lib/devicons';
import { computed, ref, watch, onUnmounted } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import { CircleX, ExternalLink, Globe, HardDrive } from '@lucide/vue';
import Spinner from '@/components/LoadingSpinner.vue';
import RelativeTime from '@/components/RelativeTime.vue';
import { useBuildLogsPanel } from '@/composables/useBuildLogsPanel';
import ResourceIcon from '@/components/ResourceIcon.vue';
import { EndpointType, ServiceStatus, type Protocol } from '@/gql/graphql';
import type { CanvasReleaseFailure, CanvasReleasePhase } from '@/composables/useCanvasReleaseStatus';
import { Status } from '@/components/ui/status';

interface Endpoint {
  host: string;
  port: number;
  protocol: Protocol;
  type: EndpointType;
  redirectTo?: string | null;
}

interface ReplicaCount {
  desired: number;
  ready: number;
}

const props = defineProps<{
  data: {
    name: string;
    sourceUrl: string;
    endpoints: Endpoint[];
    status: ServiceStatus;
    replicas: ReplicaCount;
    activeReleasePhase?: CanvasReleasePhase | null;
    activeReleaseStartedAt?: number | null;
    releaseFailure?: CanvasReleaseFailure | null;
    live?: boolean;
    volume?: { id: string; name: string; path: string; selected?: boolean; usagePercent?: number | null } | null;
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
  (e: 'select-volume'): void;
}>();

const isFromRepo = computed(() => !!props.data.sourceUrl);

const gitMask = `url("${devicon('git')}")`;

const status = computed(() => (props.data.activeReleasePhase && !props.data.live
  ? { tone: 'neutral' as const, label: 'Not live yet' }
  : serviceStatus(props.data.status)));

const logsPanel = useBuildLogsPanel();

function openFailure() {
  const failure = props.data.releaseFailure;
  if (failure?.logs) logsPanel.open(failure.logs.id, props.data.name, failure.logs.kind);
  else emit('select');
}

const shortRepoName = computed(() => {
  if (!props.data.sourceUrl) return null;
  return props.data.sourceUrl.replace('https://github.com/', '').replace(/\.git$/, '');
});

// Deploy timer
const elapsed = ref(0);
let timer: ReturnType<typeof setInterval> | null = null;

function clearTimer() {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
}

watch(() => props.data.activeReleasePhase, (phase, oldPhase) => {
  if (!phase) {
    clearTimer();
    return;
  }

  if (oldPhase) return;

  clearTimer();
  elapsed.value = props.data.activeReleaseStartedAt
    ? Math.floor((Date.now() - props.data.activeReleaseStartedAt) / 1000)
    : 0;
  timer = setInterval(() => elapsed.value++, 1000);
}, { immediate: true });

onUnmounted(clearTimer);

const deployLabel = computed(() => {
  switch (props.data.activeReleasePhase) {
    case 'queued':
      return 'Queued';
    case 'building':
      return 'Building';
    case 'deploying':
      return 'Deploying';
    case 'rollout':
      return 'Rolling out';
    default:
      return null;
  }
});

const formattedElapsed = computed(() => formatDuration(elapsed.value * 1000));

const servedEndpoints = computed(() =>
  props.data.endpoints.filter(e => e.type !== EndpointType.Internal && !e.redirectTo),
);

const primaryDomain = computed(() => {
  const custom = servedEndpoints.value.find(e => e.type === EndpointType.Custom);

  if (custom) {
    return custom.host;
  }

  const platform = servedEndpoints.value.find(e => e.type === EndpointType.Platform);

  if (platform) {
    return platform.host;
  }

  return null;
});

const extraDomainCount = computed(() => Math.max(servedEndpoints.value.length - 1, 0));

const hostUrl = computed(() => {
  if (!primaryDomain.value) {
    return null;
  }

  return `https://${primaryDomain.value}`;
});
</script>

<template>
  <div class="group w-72">

    <div
      :class="[
        'service-node group cursor-pointer rounded-xl border px-6 py-5 shadow-sm transition duration-200',
        'hover:shadow-md',
        selected ? 'border-primary shadow-md' : 'border-border',
        (deployLabel || data.releaseFailure) && 'rounded-tl-none',
      ]"
      @click="emit('select')"
    >
      <div
        v-if="deployLabel || data.releaseFailure"
        :class="[
          'absolute bottom-full -left-px flex max-w-full items-center gap-1.5 rounded-t-lg border border-b-0 px-3 py-1.5 text-xs font-medium whitespace-nowrap',
          deployLabel ? 'release-tab-progress' : 'release-tab-failed',
          selected ? 'border-primary' : 'border-border',
        ]"
      >
        <template v-if="deployLabel">
          <Spinner :size="12" class="text-status-progress" />
          <span class="text-foreground">{{ deployLabel }}</span>
          <span class="font-normal text-muted-foreground tabular-nums">{{ formattedElapsed }}</span>
        </template>
        <template v-else-if="data.releaseFailure">
          <CircleX :size="14" class="shrink-0 text-status-danger" />
          <span class="truncate text-status-danger">{{ data.releaseFailure.label }}</span>
          <span class="shrink-0 font-normal text-muted-foreground"><RelativeTime :value="data.releaseFailure.at" /></span>
          <button
            type="button"
            class="ml-1 shrink-0 text-foreground underline underline-offset-2 hover:text-muted-foreground"
            @click.stop="openFailure"
          >
            {{ data.releaseFailure.logs ? 'Logs' : 'Details' }}
          </button>
        </template>
      </div>

      <!-- Header: icon + name -->
      <div class="flex items-center gap-3">
        <ResourceIcon :kind="isFromRepo ? 'repositoryService' : 'imageService'" :size="28" />
        <span class="truncate font-semibold text-foreground">{{ data.name }}</span>
      </div>

      <!-- Domain + repo -->
      <div v-if="primaryDomain || shortRepoName" class="mt-3 space-y-1">
        <a
          v-if="primaryDomain"
          :href="hostUrl!"
          target="_blank"
          rel="noopener noreferrer"
          class="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors"
          @click.stop
        >
          <Globe :size="12" class="shrink-0" />
          <span class="truncate hover:underline">{{ primaryDomain }}</span>
          <span v-if="extraDomainCount" class="shrink-0 rounded-full bg-muted px-1.5 py-0.5 text-xs leading-none text-muted-foreground">+{{ extraDomainCount }}</span>
          <ExternalLink :size="10" class="shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" />
        </a>
        <div v-if="shortRepoName" class="flex items-center gap-1.5 text-xs text-muted-foreground">
          <span
            class="shrink-0"
            :style="{
              width: '12px',
              height: '12px',
              backgroundColor: 'currentColor',
              maskImage: gitMask,
              WebkitMaskImage: gitMask,
              maskSize: 'contain',
              WebkitMaskSize: 'contain',
              maskRepeat: 'no-repeat',
              WebkitMaskRepeat: 'no-repeat',
            }"
          />
          <span class="truncate">{{ shortRepoName }}</span>
        </div>
      </div>

      <!-- Status row -->
      <div class="mt-4 flex items-center justify-between border-t border-border/50 pt-4">
        <Status :tone="status.tone" size="sm">{{ status.label }}</Status>
      </div>

      <!-- Vue Flow handles (invisible, for potential edges) -->
      <Handle type="source" :position="Position.Bottom" class="!invisible" />
      <Handle type="target" :position="Position.Top" class="!invisible" />
    </div>

    <!-- Mounted volume, attached underneath (tucked behind the card) -->
    <div
      v-if="data.volume"
      :class="[
        'volume-attachment relative -mt-3 cursor-pointer overflow-hidden rounded-b-xl border border-t-0 transition-colors',
        data.volume.selected ? 'border-primary' : 'border-border',
      ]"
      @click.stop="emit('select-volume')"
    >
      <!-- Usage fill (background progress bar) -->
      <div
        v-if="data.volume.usagePercent != null"
        class="volume-usage-fill absolute inset-y-0 left-0 transition-[width] duration-500"
        :style="{ width: `${data.volume.usagePercent}%` }"
      />
      <!-- Content -->
      <div class="relative z-1 flex items-center gap-2.5 px-6 pb-3 pt-6">
        <HardDrive :size="16" class="shrink-0 text-muted-foreground" />
        <span class="max-w-1/2 shrink-0 truncate text-sm font-medium text-foreground">{{ data.volume.name }}</span>
        <span dir="rtl" class="ml-auto min-w-0 truncate font-mono text-xs text-muted-foreground"><bdi>{{ data.volume.path }}</bdi></span>
      </div>
    </div>

    <div class="flex flex-col-reverse relative -top-14 group-hover:translate-y-0.5 transition">
      <div v-for="i in Math.max(0, props.data.replicas.desired - 1)" :key="i" class="h-2 group-hover:h-3">
        <div class="bg-muted shadow-sm rounded-b-lg h-16 border border-muted-foreground/30"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.service-node {
  position: relative;
  z-index: 2;
  background: linear-gradient(
    to bottom,
    var(--card) 0%,
    color-mix(in oklch, var(--card) 94%, var(--muted)) 100%
  );
}

.volume-attachment {
  background: color-mix(in oklch, var(--card) 90%, var(--muted));
}

.volume-attachment:hover {
  background: color-mix(in oklch, var(--card) 86%, var(--muted));
}

.volume-usage-fill {
  background: color-mix(in oklch, var(--card) 30%, var(--muted));
}

.release-tab-progress {
  background: color-mix(in oklab, var(--card) 92%, var(--status-progress));
}

.release-tab-failed {
  background: color-mix(in oklab, var(--card) 93%, var(--status-danger));
}
</style>
