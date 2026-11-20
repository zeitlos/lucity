<script setup lang="ts">
import { computed } from 'vue';
import { CircleAlert, ShieldCheck, ShieldX } from '@lucide/vue';
import { AgentApprovalDecision, AgentApprovalStatus } from '@/gql/graphql';
import { memberName, prettyJson, type ApprovalEvent, type ToolCallEvent } from '@/lib/agentSession';
import { formatTime } from '@/lib/format';
import CodeBlock from '@/components/CodeBlock.vue';
import { Button } from '@/components/ui/button';

const props = defineProps<{
  approval: ApprovalEvent;
  call: ToolCallEvent | null;
  agentName: string;
  answering: AgentApprovalDecision | null;
}>();

const emit = defineEmits<{
  (e: 'answer', decision: AgentApprovalDecision): void;
}>();

const pending = computed(() => props.approval.approvalStatus === AgentApprovalStatus.Pending);
const allowed = computed(() => props.approval.approvalStatus === AgentApprovalStatus.Allowed);
const input = computed(() => prettyJson(props.call?.input));
</script>

<template>
  <div v-if="pending" class="overflow-hidden rounded-lg border border-status-warn/40 bg-card shadow-xs">
    <div class="flex items-start gap-3 p-3">
      <CircleAlert :size="16" class="mt-0.5 shrink-0 text-status-warn" />
      <div class="min-w-0 flex-1">
        <p v-if="call" class="text-sm font-medium text-foreground">
          {{ agentName }} wants to call <span class="font-mono">{{ call.tool }}</span><template v-if="call.connector"> on {{ call.connector }}</template>
        </p>
        <p v-else class="text-sm font-medium text-foreground">{{ approval.title }}</p>
        <p class="text-xs text-muted-foreground">Waiting since {{ formatTime(approval.time) }}</p>
      </div>
    </div>
    <div v-if="input" class="max-h-72 overflow-y-auto px-3 pb-3">
      <CodeBlock :code="input" />
    </div>
    <div class="flex justify-end gap-2 border-t bg-muted/30 px-3 py-2">
      <Button
        variant="outline"
        size="xs"
        :pending="answering === AgentApprovalDecision.Deny"
        :disabled="answering !== null"
        @click="emit('answer', AgentApprovalDecision.Deny)"
      >
        {{ answering === AgentApprovalDecision.Deny ? 'Denying…' : 'Deny' }}
      </Button>
      <Button
        size="xs"
        :pending="answering === AgentApprovalDecision.AllowOnce"
        :disabled="answering !== null"
        @click="emit('answer', AgentApprovalDecision.AllowOnce)"
      >
        {{ answering === AgentApprovalDecision.AllowOnce ? 'Allowing…' : 'Allow once' }}
      </Button>
    </div>
  </div>

  <p v-else class="flex items-center gap-1.5 text-xs text-muted-foreground">
    <ShieldCheck v-if="allowed" :size="14" class="shrink-0" />
    <ShieldX v-else :size="14" class="shrink-0 text-status-warn" />
    <span class="truncate">
      <span class="font-mono">{{ call?.tool ?? approval.title }}</span>
      · {{ allowed ? 'Allowed' : 'Denied' }} by {{ memberName(approval.decidedBy) }}
    </span>
  </p>
</template>
