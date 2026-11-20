<script setup lang="ts">
import { computed } from 'vue';
import { Ban, RotateCcw, ShieldCheck, TriangleAlert, X } from '@lucide/vue';

import type { Outcome } from '@/lib/workspaces';

const props = defineProps<{
  outcome: Outcome;
  running: boolean;
}>();

const emit = defineEmits<{
  dismiss: [];
  undo: [workspaceIDs: string[]];
  retry: [workspaceIDs: string[]];
}>();

const succeeded = computed(() => props.outcome.results.filter((result) => !result.error).map((result) => result.workspaceID));
const failed = computed(() => props.outcome.results.filter((result) => result.error));
const marked = computed(() => props.outcome.verb === 'marked');
const single = computed(() => props.outcome.results.length === 1);

function subject(count: number): string {
  if (single.value) return props.outcome.results[0]!.workspaceID;
  return count === props.outcome.results.length ? `${count} workspaces` : `${count} of ${props.outcome.results.length} workspaces`;
}

const title = computed(() => {
  if (succeeded.value.length === 0) {
    return marked.value ? `Could not mark ${subject(props.outcome.results.length)}` : `Could not clear ${subject(props.outcome.results.length)}`;
  }

  return marked.value ? `Marked ${subject(succeeded.value.length)} as abuse` : `Cleared abuse on ${subject(succeeded.value.length)}`;
});

const detail = computed(() => {
  if (succeeded.value.length === 0) return single.value ? (failed.value[0]!.error ?? '') : '';
  return marked.value ? 'Suspends on the next pass' : 'Resumes on the next pass if billing is in good standing';
});
</script>

<template>
  <div
    role="status"
    class="absolute bottom-6 left-6 z-30 w-[460px] rounded-lg bg-foreground p-1 text-background shadow-lg transition-[opacity,translate] duration-200 ease-out starting:translate-y-2 starting:opacity-0"
  >
    <div class="flex items-center gap-3 py-1.5 pr-1 pl-2">
      <span
        class="flex size-8 shrink-0 items-center justify-center rounded-sm text-white"
        :class="failed.length ? 'bg-status-warn' : marked ? 'bg-status-danger' : 'bg-background/15'"
        aria-hidden="true"
      >
        <TriangleAlert v-if="failed.length" class="size-4" />
        <Ban v-else-if="marked" class="size-4" />
        <ShieldCheck v-else class="size-4" />
      </span>
      <span class="flex min-w-0 flex-1 flex-col">
        <span class="font-medium [overflow-wrap:anywhere]">{{ title }}</span>
        <span v-if="detail" class="text-background/65">{{ detail }}</span>
      </span>
      <button
        v-if="marked && succeeded.length"
        type="button"
        class="flex h-8 shrink-0 items-center gap-1.5 rounded-sm px-2.5 font-medium hover:bg-background/10 disabled:opacity-50"
        :disabled="running"
        @click="emit('undo', succeeded)"
      >
        <RotateCcw class="size-3.5" aria-hidden="true" />
        Undo
      </button>
      <button
        v-if="failed.length && single"
        type="button"
        class="flex h-8 shrink-0 items-center gap-1.5 rounded-sm px-2.5 font-medium hover:bg-background/10 disabled:opacity-50"
        :disabled="running"
        @click="emit('retry', failed.map((result) => result.workspaceID))"
      >
        <RotateCcw class="size-3.5" aria-hidden="true" />
        Retry
      </button>
      <button
        type="button"
        class="flex size-8 shrink-0 items-center justify-center rounded-sm text-background/65 hover:bg-background/10 hover:text-background"
        aria-label="Dismiss"
        @click="emit('dismiss')"
      >
        <X class="size-4" aria-hidden="true" />
      </button>
    </div>

    <div v-if="failed.length && !single" class="mt-1 rounded-md bg-background/8 px-3 py-2">
      <ul class="flex max-h-32 flex-col gap-1 overflow-y-auto">
        <li v-for="result in failed" :key="result.workspaceID" class="flex min-w-0 gap-2">
          <span class="shrink-0 font-medium">{{ result.workspaceID }}</span>
          <span class="truncate text-background/65" :title="result.error ?? undefined">{{ result.error }}</span>
        </li>
      </ul>
      <button
        type="button"
        class="mt-2 flex h-7 items-center gap-1.5 rounded-sm bg-background px-2.5 font-medium text-foreground disabled:opacity-50"
        :disabled="running"
        @click="emit('retry', failed.map((result) => result.workspaceID))"
      >
        <RotateCcw class="size-3.5" aria-hidden="true" />
        Retry
      </button>
    </div>
  </div>
</template>
