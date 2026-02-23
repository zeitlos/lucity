<script setup lang="ts">
import { computed, ref, watch, onUnmounted } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import { ExternalLink, Github, Globe, Loader2 } from 'lucide-vue-next';
import FrameworkIcon from '@/components/FrameworkIcon.vue';
import { Badge } from '@/components/ui/badge';

const props = defineProps<{
  data: {
    name: string;
    framework?: string;
    port?: number;
    sourceUrl?: string;
    host?: string;
    ready?: boolean;
    imageTag?: string;
    replicas?: number;
    activeDeployPhase?: string | null;
    activeDeployStartedAt?: number | null;
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
}>();

const badgeVariant = computed(() => {
  if (props.data.ready === undefined) return 'secondary' as const;
  return props.data.ready ? 'default' as const : 'destructive' as const;
});

const statusLabel = computed(() => {
  if (props.data.ready === undefined) return 'Unknown';
  return props.data.ready ? 'Online' : 'Not Ready';
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

watch(() => props.data.activeDeployPhase, (phase) => {
  clearTimer();
  if (phase && phase !== 'SUCCEEDED' && phase !== 'FAILED') {
    elapsed.value = props.data.activeDeployStartedAt
      ? Math.floor((Date.now() - props.data.activeDeployStartedAt) / 1000)
      : 0;
    timer = setInterval(() => elapsed.value++, 1000);
  }
}, { immediate: true });

onUnmounted(clearTimer);

const deployLabel = computed(() => {
  switch (props.data.activeDeployPhase) {
    case 'QUEUED':
    case 'CLONING':
      return 'Initializing';
    case 'BUILDING':
    case 'PUSHING':
      return 'Building';
    case 'DEPLOYING':
      return 'Deploying';
    default:
      return null;
  }
});

const formattedElapsed = computed(() => {
  const mins = Math.floor(elapsed.value / 60);
  const secs = elapsed.value % 60;
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
});

const replicas = computed(() => props.data.replicas ?? 0);

const hostUrl = computed(() => {
  if (!props.data.host) return null;
  if (props.data.host.endsWith('.local')) {
    return `http://${props.data.host}:8880`;
  }
  return `https://${props.data.host}`;
});
</script>

<template>
  <div
    :class="[
      'service-node group cursor-pointer rounded-xl border px-6 py-5 shadow-sm transition-all duration-200',
      'hover:shadow-md',
      selected ? 'border-primary shadow-md' : 'border-border',
      replicas >= 2 && 'has-stack',
      replicas >= 3 && 'has-stack-deep',
    ]"
    style="width: 280px;"
    @click="emit('select')"
  >
    <!-- Header: icon + name -->
    <div class="flex items-center gap-3">
      <FrameworkIcon :framework="data.framework" :size="28" />
      <span class="truncate font-semibold text-foreground">{{ data.name }}</span>
    </div>

    <!-- Domain + repo -->
    <div v-if="data.host || shortRepoName" class="mt-3 space-y-1">
      <a
        v-if="data.host"
        :href="hostUrl!"
        target="_blank"
        rel="noopener noreferrer"
        class="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors"
        @click.stop
      >
        <Globe :size="12" class="shrink-0" />
        <span class="truncate hover:underline">{{ data.host }}</span>
        <ExternalLink :size="10" class="shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" />
      </a>
      <div v-if="shortRepoName" class="flex items-center gap-1.5 text-xs text-muted-foreground">
        <Github :size="12" class="shrink-0" />
        <span class="truncate">{{ shortRepoName }}</span>
      </div>
    </div>

    <!-- Status row -->
    <div class="mt-4 flex items-center justify-between border-t border-border/50 pt-4">
      <Badge :variant="badgeVariant" class="text-[0.65rem]">{{ statusLabel }}</Badge>
      <span v-if="deployLabel" class="flex items-center gap-1.5 text-[0.65rem] text-muted-foreground">
        <Loader2 :size="12" class="animate-spin text-primary" />
        {{ deployLabel }} ({{ formattedElapsed }})
      </span>
    </div>

    <!-- Vue Flow handles (invisible, for potential edges) -->
    <Handle type="source" :position="Position.Bottom" class="!invisible" />
    <Handle type="target" :position="Position.Top" class="!invisible" />
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

/* First stacked card (2+ replicas) */
.has-stack::before {
  content: '';
  position: absolute;
  inset: 0;
  z-index: -1;
  border-radius: inherit;
  border: 1px solid var(--border);
  background: var(--card);
  transform: translateY(6px) scale(0.97);
  opacity: 0.7;
}

/* Second stacked card (3+ replicas) */
.has-stack-deep::after {
  content: '';
  position: absolute;
  inset: 0;
  z-index: -2;
  border-radius: inherit;
  border: 1px solid var(--border);
  background: var(--card);
  transform: translateY(12px) scale(0.94);
  opacity: 0.4;
}
</style>
