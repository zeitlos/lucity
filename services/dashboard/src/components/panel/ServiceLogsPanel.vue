<script setup lang="ts">
import { useEscapeToClose } from '@/composables/useEscapeToClose';
import { ref, computed, watch, nextTick } from 'vue';
import { X, Pause, Play, Eraser, CircleAlert, RotateCw } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import Spinner from '@/components/LoadingSpinner.vue';
import LogLine from '@/components/LogLine.vue';
import { useServiceLogs } from '@/composables/useServiceLogs';
import IconButton from '@/components/IconButton.vue';

const props = defineProps<{
  serviceId: string;
  serviceName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

useEscapeToClose(() => emit('close'));

const serviceIdRef = computed(() => props.serviceId);
const enabled = ref(true);

const { lines, isActive, error, clear, stop, restart, retry } = useServiceLogs(
  serviceIdRef,
  enabled,
);

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

function togglePause() {
  paused.value = !paused.value;
  if (paused.value) {
    stop();
  } else {
    enabled.value = true;
    restart();
  }
}

function clearLogs() {
  clear();
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
      </div>

      <div class="flex items-center gap-1">
        <IconButton
          :label="paused ? 'Resume' : 'Pause'"
          class="text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
          @click="togglePause"
        >
          <Pause
            v-if="!paused"
          />
          <Play
            v-else
          />
        </IconButton>
        <IconButton
          label="Clear logs"
          class="text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
          @click="clearLogs"
        >
          <Eraser />
        </IconButton>
        <IconButton
          label="Close" shortcut="Esc"
          class="text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
          @click="emit('close')"
        >
          <X />
        </IconButton>
      </div>
    </div>

    <!-- Log output -->
    <div
      ref="logContainer"
      class="flex-1 overflow-auto p-4 font-mono text-xs leading-relaxed text-zinc-300"
      @scroll="handleScroll"
    >
      <div v-if="error && lines.length === 0" class="flex flex-col items-start gap-2 font-sans">
        <p class="flex items-center gap-2 text-sm font-medium text-status-danger">
          <CircleAlert :size="14" />
          Couldn't load logs
        </p>
        <p class="text-zinc-400">{{ error }}</p>
        <Button variant="outline" size="xs" class="border-zinc-700 bg-transparent text-zinc-200" @click="retry">
          <RotateCw />
          Try again
        </Button>
      </div>

      <div
        v-else-if="lines.length === 0 && !isActive"
        class="flex items-center gap-2 text-zinc-500"
      >
        <Spinner :size="12" />
        <span>Waiting for logs…</span>
      </div>

      <div
        v-for="(entry, idx) in lines"
        :key="idx"
      >
        <span class="select-none pr-3 text-zinc-600">{{ String(idx + 1).padStart(4, ' ') }}</span>
        <LogLine :text="entry.line" />
      </div>

      <div
        v-if="isActive && lines.length > 0 && !paused"
        class="mt-2 flex items-center gap-2 text-zinc-500"
      >
        <Spinner
          :size="12"
          class="animate-spin"
        />
      </div>
    </div>
  </div>
</template>
