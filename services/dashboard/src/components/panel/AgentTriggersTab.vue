<script setup lang="ts">
import { computed, ref } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { Clock, Plus, Trash2, Webhook, Zap } from '@lucide/vue';
import { graphql } from '@/gql';
import { AgentStatus, AgentTriggerSource } from '@/gql/graphql';
import type { Agent, AgentTrigger } from '@/lib/agents';
import { brandIcon } from '@/lib/brandMarks';
import { describeCron } from '@/lib/schedule';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import EmptyState from '@/components/EmptyState.vue';
import IconButton from '@/components/IconButton.vue';
import QueryError from '@/components/QueryError.vue';
import RelativeTime from '@/components/RelativeTime.vue';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { failureToast } from '@/components/ui/sonner';
import AgentTriggerDialog from './AgentTriggerDialog.vue';

const SetAgentTriggerEnabledDocument = graphql(`
  mutation SetAgentTriggerEnabled($trigger: AgentTriggerID!, $enabled: Boolean!) {
    setAgentTriggerEnabled(trigger: $trigger, enabled: $enabled) {
      id
      enabled
    }
  }
`);

const DeleteAgentTriggerDocument = graphql(`
  mutation DeleteAgentTrigger($trigger: AgentTriggerID!) {
    deleteAgentTrigger(trigger: $trigger) {
      id
      triggers {
        id
      }
    }
  }
`);

const AgentTriggerRunsDocument = graphql(`
  query AgentTriggerRuns($agent: AgentID!) {
    agent(id: $agent) {
      id
      triggers {
        id
        lastRunAt
      }
    }
  }
`);

const props = defineProps<{
  agent: Agent;
}>();

const starting = computed(() => props.agent.status === AgentStatus.Starting);

const { result: runsResult, error: runsError, refetch: refetchRuns } = useQuery(
  AgentTriggerRunsDocument,
  () => ({ agent: props.agent.id }),
  () => ({ enabled: !starting.value && props.agent.triggers.length > 0, pollInterval: 30000 }),
);

const lastRuns = computed(() => new Map((runsResult.value?.agent.triggers ?? []).map(trigger => [trigger.id, trigger.lastRunAt])));

const dialogOpen = ref(false);
const editing = ref<AgentTrigger | null>(null);
const toggling = ref<string | null>(null);
const deleting = ref<string | null>(null);

const { mutate: setEnabled } = useMutation(SetAgentTriggerEnabledDocument);
const { mutate: deleteTrigger } = useMutation(DeleteAgentTriggerDocument);

function openNew() {
  editing.value = null;
  dialogOpen.value = true;
}

function openEdit(trigger: AgentTrigger) {
  editing.value = trigger;
  dialogOpen.value = true;
}

const GitHubIcon = brandIcon('github');

function when(trigger: AgentTrigger) {
  if (trigger.source === AgentTriggerSource.Webhook) return 'Webhook';
  if (trigger.source === AgentTriggerSource.Github) return trigger.repository ?? 'GitHub';
  return `${describeCron(trigger.schedule ?? '') ?? trigger.schedule} · ${trigger.timeZone}`;
}

async function toggle(trigger: AgentTrigger, enabled: boolean) {
  toggling.value = trigger.id;
  try {
    await setEnabled({ trigger: trigger.id, enabled });
  } catch (e: unknown) {
    failureToast(`Failed to turn "${trigger.name}" ${enabled ? 'on' : 'off'}`, e);
  } finally {
    toggling.value = null;
  }
}

async function remove(trigger: AgentTrigger) {
  deleting.value = trigger.id;
  try {
    await deleteTrigger({ trigger: trigger.id });
  } catch (e: unknown) {
    failureToast('Failed to delete trigger', e);
  } finally {
    deleting.value = null;
  }
}
</script>

<template>
  <section class="space-y-4">
    <div class="flex items-start justify-between gap-4">
      <p class="text-sm text-muted-foreground">
        Each trigger starts a session with {{ agent.name }} on its own<template v-if="agent.chat"> and posts it in {{ agent.chat.channel.name }}</template>.
      </p>
      <Button variant="outline" size="xs" class="shrink-0" @click="openNew">
        <Plus />
        New trigger
      </Button>
    </div>

    <EmptyState
      v-if="agent.triggers.length === 0"
      size="compact"
      :icon="Zap"
      title="No triggers yet"
      description="Start sessions on a schedule, when a webhook is called, or on GitHub events."
    />
    <QueryError v-if="runsError && !runsResult && agent.triggers.length > 0" title="Couldn't load when the triggers last ran" :error="runsError" @retry="refetchRuns()" />

    <div v-if="agent.triggers.length > 0" class="-mx-4 divide-y border-y">
      <div v-for="trigger in agent.triggers" :key="trigger.id" class="flex items-center gap-3 py-2.5 pr-3 pl-4">
        <span class="inline-flex size-8 shrink-0 items-center justify-center rounded-md border bg-muted text-foreground">
          <Webhook v-if="trigger.source === AgentTriggerSource.Webhook" :size="16" />
          <component :is="GitHubIcon" v-else-if="trigger.source === AgentTriggerSource.Github && GitHubIcon" class="size-4" />
          <Clock v-else :size="16" />
        </span>
        <button type="button" class="min-w-0 flex-1 text-left" @click="openEdit(trigger)">
          <p class="truncate text-sm font-medium text-foreground">{{ trigger.name }}</p>
          <p class="truncate text-xs text-muted-foreground">
            {{ when(trigger) }}<template v-if="lastRuns.has(trigger.id)"> ·
              <template v-if="lastRuns.get(trigger.id)">last run <RelativeTime :value="lastRuns.get(trigger.id)!" /></template>
              <template v-else>not run yet</template>
            </template>
          </p>
        </button>
        <Switch
          :model-value="trigger.enabled"
          :disabled="toggling === trigger.id"
          class="shrink-0 data-[state=unchecked]:bg-border"
          :aria-label="`${trigger.name} on`"
          @update:model-value="(value: boolean) => toggle(trigger, value)"
        />
        <ConfirmDialog
          :title="`Delete trigger &quot;${trigger.name}&quot;?`"
          description="It stops starting sessions. The sessions it started stay."
          confirm-label="Delete trigger"
          pending-label="Deleting…"
          :pending="deleting === trigger.id"
          destructive
          @confirm="remove(trigger)"
        >
          <template #trigger>
            <IconButton :label="`Delete ${trigger.name}`" size="icon-xs" class="shrink-0">
              <Trash2 />
            </IconButton>
          </template>
        </ConfirmDialog>
      </div>
    </div>

    <AgentTriggerDialog v-model:open="dialogOpen" :agent="agent" :trigger="editing" />
  </section>
</template>
