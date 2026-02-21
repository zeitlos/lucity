<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { X, Loader2, Trash2, Pause, Play } from 'lucide-vue-next';
import { onKeyStroke } from '@vueuse/core';
import { useServiceLogs } from '@/composables/useServiceLogs';
import { Button } from '@/components/ui/button';

const props = defineProps<{
  projectId: string;
  serviceName: string;
  environment: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

onKeyStroke('Escape', () => emit('close'));

const projectIdRef = computed(() => props.projectId);
const serviceRef = computed(() => props.serviceName);
const envRef = computed(() => props.environment as string | null);
const enabled = ref(true);

const { lines, isActive, clear, stop, restart } = useServiceLogs(
  projectIdRef,
  serviceRef,
  envRef,
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
        <span class="text-xs text-zinc-500">
          {{ environment }}
        </span>
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
          @click="clearLogs"
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
        v-if="lines.length === 0 && !isActive"
        class="flex items-center gap-2 text-zinc-500"
      >
        <Loader2
          :size="12"
          class="animate-spin"
        />
        <span>Waiting for logs...</span>
      </div>

      <div
        v-for="(entry, idx) in lines"
        :key="idx"
      >
        <span class="select-none pr-3 text-zinc-600">{{ String(idx + 1).padStart(4, ' ') }}</span>
        <span class="whitespace-pre-wrap break-all">{{ entry.line }}</span>
      </div>

      <div
        v-if="isActive && lines.length > 0 && !paused"
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
