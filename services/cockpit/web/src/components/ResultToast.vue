<script setup lang="ts">
import { computed } from 'vue';
import { X } from '@lucide/vue';

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
</script>

<template>
  <div role="status" class="absolute bottom-6 left-6 z-30 flex w-[440px] flex-col gap-2 rounded-lg bg-foreground px-4 py-3 text-background shadow-lg">
    <div class="flex items-start gap-3">
      <span class="font-medium">{{ succeeded.length }} of {{ outcome.results.length }} {{ outcome.verb }}</span>
      <button type="button" class="ml-auto opacity-80 hover:opacity-100" aria-label="Dismiss" @click="emit('dismiss')">
        <X class="size-4" aria-hidden="true" />
      </button>
    </div>
    <ul v-if="failed.length" class="flex flex-col gap-1 opacity-90">
      <li v-for="result in failed" :key="result.workspaceID">
        <span class="font-medium">{{ result.workspaceID }}</span> {{ result.error }}
      </li>
    </ul>
    <div class="flex gap-2">
      <button
        v-if="outcome.verb === 'marked' && succeeded.length"
        type="button"
        class="h-8 rounded-sm border border-background/30 px-3 disabled:opacity-50"
        :disabled="running"
        @click="emit('undo', succeeded)"
      >
        Undo {{ succeeded.length }}
      </button>
      <button
        v-if="failed.length"
        type="button"
        class="h-8 rounded-sm border border-background/30 px-3 disabled:opacity-50"
        :disabled="running"
        @click="emit('retry', failed.map((result) => result.workspaceID))"
      >
        Retry {{ failed.length === 1 ? failed[0]!.workspaceID : failed.length }}
      </button>
    </div>
  </div>
</template>
