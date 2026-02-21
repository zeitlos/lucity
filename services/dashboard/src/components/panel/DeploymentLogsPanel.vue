<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted } from 'vue';
import { X, Loader2 } from 'lucide-vue-next';
import { onKeyStroke } from '@vueuse/core';
import { useDeployLogs } from '@/composables/useDeployLogs';
import { useDeploy } from '@/composables/useDeploy';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

const props = defineProps<{
  deployId: string;
  serviceName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

onKeyStroke('Escape', () => emit('close'));

const deployIdRef = computed(() => props.deployId);
const { lines, isActive } = useDeployLogs(deployIdRef);
const deploy = useDeploy();

const logContainer = ref<HTMLElement | null>(null);
const userScrolled = ref(false);

function handleScroll() {
  if (!logContainer.value) return;
  const el = logContainer.value;
  const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
  userScrolled.value = !atBottom;
}

watch(lines, async () => {
  if (userScrolled.value) return;
  await nextTick();
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight;
  }
}, { deep: true });

const isTerminal = computed(() =>
  deploy.phase === 'SUCCEEDED' || deploy.phase === 'FAILED'
);

onUnmounted(() => {
  // Composable cleanup handled by Vue reactivity
});
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
          v-if="deploy.phase"
          :variant="deploy.phase === 'SUCCEEDED' ? 'default' : deploy.phase === 'FAILED' ? 'destructive' : 'secondary'"
          class="text-xs"
        >
          <Loader2
            v-if="deploy.isDeploying"
            :size="10"
            class="mr-1 animate-spin"
          />
          {{ deploy.phase }}
        </Badge>
      </div>

      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7 text-zinc-400 hover:text-zinc-200"
        @click="emit('close')"
      >
        <X :size="16" />
      </Button>
    </div>

    <!-- Log output -->
    <div
      ref="logContainer"
      class="flex-1 overflow-auto p-4 font-mono text-xs leading-relaxed text-zinc-300"
      @scroll="handleScroll"
    >
      <div v-if="lines.length === 0 && !isTerminal" class="flex items-center gap-2 text-zinc-500">
        <Loader2 :size="12" class="animate-spin" />
        <span>Waiting for logs...</span>
      </div>

      <div v-for="(line, idx) in lines" :key="idx">
        <span class="select-none pr-3 text-zinc-600">{{ String(idx + 1).padStart(4, ' ') }}</span>
        <span class="whitespace-pre-wrap break-all">{{ line }}</span>
      </div>

      <div
        v-if="isActive && !isTerminal && lines.length > 0"
        class="mt-2 flex items-center gap-2 text-zinc-500"
      >
        <Loader2 :size="12" class="animate-spin" />
      </div>
    </div>
  </div>
</template>
