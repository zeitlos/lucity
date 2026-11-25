<script setup lang="ts">
import { computed, useTemplateRef } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import { Bot, CircleAlert, Shield, Zap } from '@lucide/vue';
import { AgentStatus, type ChatAppKind, type Harness } from '@/gql/graphql';
import { agentStatus } from '@/lib/status';
import { chatAppKindLabel, harnessLabel } from '@/lib/agents';
import { brandMark } from '@/lib/brandMarks';
import { darkTooltipClass } from '@/lib/tooltip';
import BrandTile from '@/components/BrandTile.vue';
import { Status } from '@/components/ui/status';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useGridHeight } from '@/composables/useGridHeight';

const props = defineProps<{
  data: {
    name: string;
    harness: Harness;
    status: AgentStatus;
    connectors: string[];
    chat: { kind: ChatAppKind; channel: string } | null;
    triggers: string[];
    allowedHosts: number;
    privateNetwork: boolean;
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
}>();

useGridHeight(useTemplateRef<HTMLElement>('card'));

const status = computed(() => agentStatus(props.data.status));
const waiting = computed(() => props.data.status === AgentStatus.WaitingForApproval);
const harnessMark = computed(() => brandMark(harnessLabel(props.data.harness)));
const chatMark = computed(() => (props.data.chat ? brandMark(chatAppKindLabel(props.data.chat.kind)) : null));

const network = computed(() => {
  const hosts = props.data.allowedHosts;
  if (hosts === 0) return props.data.privateNetwork ? 'Private network' : null;
  const allowed = `${hosts} allowed ${hosts === 1 ? 'host' : 'hosts'}`;
  return props.data.privateNetwork ? `${allowed}, private network` : allowed;
});
</script>

<template>
  <div class="w-72">
    <div
      :class="[
        'agent-node group cursor-pointer rounded-xl border px-6 py-5 shadow-sm transition duration-200 flex flex-col',
        'hover:shadow-md',
        selected ? 'border-primary shadow-md' : 'border-border',
        waiting && 'rounded-tl-none',
      ]"
      ref="card"
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
        <svg
          v-if="harnessMark"
          :viewBox="harnessMark.viewBox ?? '0 0 24 24'"
          class="size-7 shrink-0"
          :style="{ color: harnessMark.background }"
          fill="currentColor"
          role="img"
          :aria-label="harnessLabel(data.harness)"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path :d="harnessMark.path" />
        </svg>
        <Bot v-else :size="28" class="shrink-0 text-foreground" />
        <span class="truncate font-semibold text-foreground">{{ data.name }}</span>
      </div>

      <div v-if="data.chat || data.triggers.length || network" class="mt-3 space-y-1 text-xs text-muted-foreground">
        <div v-if="data.chat" class="flex items-center gap-1.5">
          <svg
            v-if="chatMark"
            :viewBox="chatMark.viewBox ?? '0 0 24 24'"
            class="size-3 shrink-0"
            fill="currentColor"
            role="img"
            :aria-label="chatAppKindLabel(data.chat.kind)"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path :d="chatMark.path" />
          </svg>
          <span class="truncate">{{ data.chat.channel }}</span>
        </div>
        <div v-if="data.triggers.length" class="flex items-center gap-1.5">
          <Zap :size="12" class="shrink-0" />
          <span class="truncate">{{ data.triggers[0] }}</span>
          <span v-if="data.triggers.length > 1" class="shrink-0 rounded-full bg-muted px-1.5 py-0.5 text-xs leading-none text-muted-foreground">+{{ data.triggers.length - 1 }}</span>
        </div>
        <div v-if="network" class="flex items-center gap-1.5">
          <Shield :size="12" class="shrink-0" />
          <span class="truncate">{{ network }}</span>
        </div>
      </div>

      <div class="grow basis-3" />
      <div class="mt-1 flex items-center justify-between gap-3 border-t border-border/50 pt-4">
        <Status :tone="status.tone" :glyph="status.glyph" size="sm">{{ status.label }}</Status>
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
