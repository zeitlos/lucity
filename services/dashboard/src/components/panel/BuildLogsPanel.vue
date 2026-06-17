<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { X, Loader2, Trash2, Pause, Play, AlertCircle } from 'lucide-vue-next';
import { onKeyStroke } from '@vueuse/core';
import { useBuildLogs } from '@/composables/useBuildLogs';
import { BuildStatus } from '@/gql/graphql';
import { useDeploy } from '@/composables/useDeploy';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

const props = defineProps<{
  buildId: string;
  serviceName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

onKeyStroke('Escape', () => emit('close'));

const buildIdRef = computed(() => props.buildId);
const { lines, isActive, error, clear, stop, restart } = useBuildLogs(buildIdRef);
const deploy = useDeploy();

const logContainer = ref<HTMLElement | null>(null);
const userScrolled = ref(false);
const paused = ref(false);

function handleScroll() {
  if (!logContainer.value) return;
  const el = logContainer.value;
  const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
  userScrolled.value = !atBottom;
}

watch(lines, async () => {
  if (userScrolled.value || paused.value) return;
  await nextTick();
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight;
  }
}, { deep: true });

const isTerminal = computed(() =>
  deploy.status === BuildStatus.Succeeded
    || deploy.status === BuildStatus.Failed
    || deploy.status === BuildStatus.Cancelled
);

function togglePause() {
  paused.value = !paused.value;
  if (paused.value) {
    stop();
  } else {
    restart();
  }
}
</script>

<template>
  <div class="flex h-full flex-col rounded-lg border bg-zinc-950 shadow-2xl">
    <!-- Header -->
    <div class="flex shrink-0 items-center justify-between border-b border-zinc-800 px-4 py-3">
      <div class="flex items-center gap-3">
        <h2 class="text-sm font-semibold text-zinc-200">
          {{ serviceName }}
        </h2>
        <Badge
          v-if="deploy.status"
          :variant="deploy.status === BuildStatus.Succeeded ? 'default' : deploy.status === BuildStatus.Failed ? 'destructive' : 'secondary'"
          class="text-xs"
        >
          <Loader2
            v-if="deploy.isDeploying"
            :size="10"
            class="mr-1 animate-spin"
          />
          {{ deploy.status }}
        </Badge>
      </div>

      <div class="flex items-center gap-1">
        <Button
          variant="ghost"
          size="icon"
          class="h-6 w-6 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
          @click="togglePause"
        >
          <Pause
            v-if="!paused"
            :size="12"
          />
          <Play
            v-else
            :size="12"
          />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          class="h-6 w-6 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
          @click="clear"
        >
          <Trash2 :size="12" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          class="h-7 w-7 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
          @click="emit('close')"
        >
          <X :size="16" />
        </Button>
      </div>
    </div>

    <!-- Log output -->
    <div
      ref="logContainer"
      class="flex-1 overflow-auto p-4 font-mono text-xs leading-relaxed text-zinc-300"
      @scroll="handleScroll"
    >
      <div
        v-if="error"
        class="flex items-start gap-2 rounded-md border border-red-900/40 bg-red-950/30 px-3 py-2.5 text-red-300"
      >
        <AlertCircle :size="13" class="mt-0.5 shrink-0" />
        <div class="min-w-0 space-y-0.5">
          <p class="font-sans text-xs font-medium">Failed to load logs</p>
          <p class="break-words font-mono text-[11px] text-red-400/80">{{ error }}</p>
        </div>
      </div>

      <div
        v-else-if="lines.length === 0 && !isTerminal"
        class="flex items-center gap-2 text-zinc-500"
      >
        <Loader2
          :size="12"
          class="animate-spin"
        />
        <span>Waiting for logs...</span>
      </div>

      <div
        v-for="(line, idx) in lines"
        :key="idx"
      >
        <span class="select-none pr-3 text-zinc-600">{{ String(idx + 1).padStart(4, ' ') }}</span>
        <span class="whitespace-pre-wrap break-all">{{ line }}</span>
      </div>

      <div
        v-if="isActive && !isTerminal && lines.length > 0 && !paused"
        class="mt-2 flex items-center gap-2 text-zinc-500"
      >
        <Loader2
          :size="12"
          class="animate-spin"
        />
      </div>
    </div>
  </div>
</template>
