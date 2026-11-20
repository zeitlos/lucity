<script setup lang="ts">
import { ref } from 'vue';
import { ChevronRight, Wrench } from '@lucide/vue';
import { AgentToolCallStatus } from '@/gql/graphql';
import { prettyJson, type ToolCallEvent } from '@/lib/agentSession';
import BrandTile from '@/components/BrandTile.vue';
import CodeBlock from '@/components/CodeBlock.vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { Status } from '@/components/ui/status';

defineProps<{
  calls: ToolCallEvent[];
}>();

const expanded = ref(new Set<string>());

function toggle(callId: string) {
  const next = new Set(expanded.value);
  if (next.has(callId)) next.delete(callId);
  else next.add(callId);
  expanded.value = next;
}
</script>

<template>
  <div class="divide-y overflow-hidden rounded-lg border bg-card">
    <div v-for="call in calls" :key="call.callId">
      <button
        type="button"
        class="flex w-full items-center gap-2 px-3 py-2 text-left text-xs transition-colors hover:bg-muted/50"
        :aria-expanded="expanded.has(call.callId)"
        @click="toggle(call.callId)"
      >
        <BrandTile v-if="call.connector" :name="call.connector" size="sm" />
        <span v-else class="flex size-5 shrink-0 items-center justify-center text-muted-foreground">
          <Wrench :size="14" />
        </span>
        <span v-if="call.connector" class="shrink-0 font-medium text-foreground">{{ call.connector }}</span>
        <span class="shrink-0 font-mono text-foreground">{{ call.tool }}</span>
        <span class="min-w-0 flex-1 truncate text-muted-foreground">
          <template v-if="call.title !== call.tool">{{ call.title }}</template>
        </span>
        <Spinner v-if="call.status === AgentToolCallStatus.Running" :size="12" class="shrink-0 text-status-progress" />
        <Status v-else-if="call.status === AgentToolCallStatus.Failed" variant="text" tone="danger">Failed</Status>
        <Status v-else-if="call.status === AgentToolCallStatus.Denied" variant="text" tone="warn">Denied</Status>
        <ChevronRight
          :size="14"
          class="shrink-0 text-muted-foreground transition-transform"
          :class="expanded.has(call.callId) && 'rotate-90'"
        />
      </button>
      <div v-if="expanded.has(call.callId)" class="max-h-96 space-y-2 overflow-y-auto border-t bg-muted/20 p-3">
        <CodeBlock v-if="call.input" title="Input" :code="prettyJson(call.input)" />
        <CodeBlock v-if="call.output" title="Output" :code="prettyJson(call.output)" />
        <p v-if="!call.input && !call.output" class="text-xs text-muted-foreground">No input or output.</p>
      </div>
    </div>
  </div>
</template>
