<script setup lang="ts">
import { computed, ref, watch, onUnmounted } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import { ExternalLink, Globe, Loader2, Container, FolderGit2, HardDrive } from '@lucide/vue';
import GithubIcon from '@/components/GithubIcon.vue';
import { BuildStatus, EndpointType, ServiceStatus, type Protocol } from '@/gql/graphql';
import { Status } from '@/components/ui/status';

interface Endpoint {
  host: string;
  port: number;
  protocol: Protocol;
  type: EndpointType;
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
    activeBuildStatus?: BuildStatus | null;
    activeBuildStartedAt?: number | null;
    volume?: { id: string; name: string; path: string; selected?: boolean; usagePercent?: number | null } | null;
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
  (e: 'select-volume'): void;
}>();

const isFromRepo = computed(() => !!props.data.sourceUrl);

const statusTone = computed(() => {
  switch (props.data.status) {
    case ServiceStatus.Healthy:
      return 'ok' as const;
    case ServiceStatus.Failed:
      return 'danger' as const;
    case ServiceStatus.Deploying:
      return 'warn' as const;
    default:
      return 'neutral' as const;
  }
});

const statusLabel = computed(() => {
  switch (props.data.status) {
    case ServiceStatus.Healthy:
      return 'Online';
    case ServiceStatus.Degraded:
      return 'Degraded';
    case ServiceStatus.Deploying:
      return 'Deploying';
    case ServiceStatus.Failed:
      return 'Failed';
    case ServiceStatus.Stopped:
      return 'Stopped';
    default:
      return 'Unknown';
  }
});

const shortRepoName = computed(() => {
  if (!props.data.sourceUrl) return null;
  return props.data.sourceUrl.replace('https://github.com/', '');
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

watch(() => props.data.activeBuildStatus, (status) => {
  clearTimer();
  if (status === BuildStatus.Queued || status === BuildStatus.Running) {
    elapsed.value = props.data.activeBuildStartedAt
      ? Math.floor((Date.now() - props.data.activeBuildStartedAt) / 1000)
      : 0;
    timer = setInterval(() => elapsed.value++, 1000);
  }
}, { immediate: true });

onUnmounted(clearTimer);

const deployLabel = computed(() => {
  switch (props.data.activeBuildStatus) {
    case BuildStatus.Queued:
      return 'Queued';
    case BuildStatus.Running:
      return 'Building';
    default:
      return null;
  }
});

const formattedElapsed = computed(() => {
  const mins = Math.floor(elapsed.value / 60);
  const secs = elapsed.value % 60;
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
});

const primaryDomain = computed(() => {
  const custom = props.data.endpoints.find(e => e.type === EndpointType.Custom);

  if (custom) {
    return custom.host;
  }

  const platform = props.data.endpoints.find(e => e.type === EndpointType.Platform);

  if (platform) {
    return platform.host;
  }

  return null;
});

const extraDomainCount = computed(() => {
  const total = props.data.endpoints.filter(e => e.type !== EndpointType.Internal).length;
  return total > 1 ? total - 1 : 0;
});

const hostUrl = computed(() => {
  if (!primaryDomain.value) {
    return null;
  }

  return `https://${primaryDomain.value}`;
});
</script>

<template>
  <div class="group">

    <div
      :class="[
        'service-node group cursor-pointer rounded-xl border px-6 py-5 shadow-sm transition duration-200 w-72',
        'hover:shadow-md',
        selected ? 'border-primary shadow-md' : 'border-border',
      ]"
      @click="emit('select')"
    >
      <!-- Header: icon + name -->
      <div class="flex items-center gap-3">
        <GithubIcon v-if="isFromRepo" :size="28" class="shrink-0" />
        <Container v-else :size="28" />
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
          <span v-if="extraDomainCount" class="shrink-0 rounded-full bg-muted px-1.5 py-0.5 text-[0.6rem] leading-none text-muted-foreground">+{{ extraDomainCount }}</span>
          <ExternalLink :size="10" class="shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" />
        </a>
        <div v-if="shortRepoName" class="flex items-center gap-1.5 text-xs text-muted-foreground">
          <FolderGit2 :size="12" class="shrink-0" />
          <span class="truncate">{{ shortRepoName }}</span>
        </div>
      </div>

      <!-- Status row -->
      <div class="mt-4 flex items-center justify-between border-t border-border/50 pt-4">
        <Status :tone="statusTone" class="text-[0.65rem]">{{ statusLabel }}</Status>
        <span v-if="deployLabel" class="flex items-center gap-1.5 text-[0.65rem] text-muted-foreground">
          <Loader2 :size="12" class="animate-spin text-primary" />
          {{ deployLabel }} ({{ formattedElapsed }})
        </span>
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
      <div class="relative z-[1] flex items-center gap-2.5 px-6 pb-3 pt-6">
        <HardDrive :size="16" class="shrink-0 text-muted-foreground" />
        <span class="truncate text-sm font-medium text-foreground">{{ data.volume.name }}</span>
        <span class="ml-auto shrink-0 font-mono text-xs text-muted-foreground">{{ data.volume.path }}</span>
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
  z-index: 1;
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
</style>
