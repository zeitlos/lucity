<script setup lang="ts">
import { computed } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import { Bot, CircleAlert } from '@lucide/vue';
import { AgentStatus, type Harness } from '@/gql/graphql';
import { agentStatus } from '@/lib/status';
import { harnessLabel } from '@/lib/agents';
import { darkTooltipClass } from '@/lib/tooltip';
import BrandTile from '@/components/BrandTile.vue';
import { Status } from '@/components/ui/status';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

const props = defineProps<{
  data: {
    name: string;
    harness: Harness;
    status: AgentStatus;
    connectors: string[];
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
}>();

const status = computed(() => agentStatus(props.data.status));
const waiting = computed(() => props.data.status === AgentStatus.WaitingForApproval);
</script>

<template>
  <div class="w-72">
    <div
      :class="[
        'agent-node group cursor-pointer rounded-xl border px-6 py-5 shadow-sm transition duration-200',
        'hover:shadow-md',
        selected ? 'border-primary shadow-md' : 'border-border',
        waiting && 'rounded-tl-none',
      ]"
      @click="emit('select')"
    >
      <div
        v-if="waiting"
        :class="[
          'agent-tab-warn absolute bottom-full -left-px flex max-w-full items-center gap-1.5 rounded-t-lg border border-b-0 px-3 py-1.5 text-xs font-medium whitespace-nowrap',
          selected ? 'border-primary' : 'border-border',
        ]"
      >
        <CircleAlert :size="14" class="shrink-0 text-status-warn" />
        <span class="truncate text-status-warn">Waiting for approval</span>
        <button
          type="button"
          class="ml-1 shrink-0 text-foreground underline underline-offset-2 hover:text-muted-foreground"
          @click.stop="emit('select')"
        >
          Review
        </button>
      </div>

      <div class="flex items-center gap-3">
        <Bot :size="28" class="shrink-0 text-foreground" />
        <span class="truncate font-semibold text-foreground">{{ data.name }}</span>
      </div>

      <div class="mt-3 text-xs text-muted-foreground">
        {{ harnessLabel(data.harness) }}
      </div>

      <div class="mt-4 flex items-center justify-between gap-3 border-t border-border/50 pt-4">
        <Status :tone="status.tone" size="sm">{{ status.label }}</Status>
        <div v-if="data.connectors.length" class="flex min-w-0 items-center -space-x-1">
          <Tooltip v-for="connector in data.connectors" :key="connector">
            <TooltipTrigger as-child>
              <span class="rounded-sm ring-2 ring-card">
                <BrandTile :name="connector" size="sm" />
              </span>
            </TooltipTrigger>
            <TooltipContent side="top" :class="darkTooltipClass">{{ connector }}</TooltipContent>
          </Tooltip>
        </div>
      </div>

      <Handle type="source" :position="Position.Bottom" class="!invisible" />
      <Handle type="target" :position="Position.Top" class="!invisible" />
    </div>
  </div>
</template>

<style scoped>
.agent-node {
  position: relative;
  z-index: 2;
  background: linear-gradient(
    to bottom,
    var(--card) 0%,
    color-mix(in oklch, var(--card) 94%, var(--muted)) 100%
  );
}

.agent-tab-warn {
  background: color-mix(in oklab, var(--card) 92%, var(--status-warn));
}
</style>
