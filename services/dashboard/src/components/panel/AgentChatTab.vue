<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { Plus } from '@lucide/vue';
import { graphql } from '@/gql';
import { AgentSessionStatus, Effort } from '@/gql/graphql';
import { memberName } from '@/lib/agentSession';
import { missingModelsHint, type Agent, type AgentModelOption } from '@/lib/agents';
import { agentSessionStatus } from '@/lib/status';
import QueryError from '@/components/QueryError.vue';
import RelativeTime from '@/components/RelativeTime.vue';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Skeleton } from '@/components/ui/skeleton';
import { Status } from '@/components/ui/status';
import AgentSessionView from './AgentSessionView.vue';

const AgentSessionsDocument = graphql(`
  query AgentSessions($agent: AgentID!) {
    agent(id: $agent) {
      id
      sessions {
        id
        title
        startedAt
        status
        effort
        startedBy { id name email }
        model { name provider { id name } }
      }
    }
  }
`);

const AgentModelsDocument = graphql(`
  query AgentModels($agent: AgentID!) {
    agentModels(agent: $agent) {
      name
      provider { id name }
    }
  }
`);

const props = defineProps<{
  agent: Agent;
}>();

const { result, loading, error, refetch } = useQuery(
  AgentSessionsDocument,
  () => ({ agent: props.agent.id }),
  { pollInterval: 5000 },
);

const { result: modelsResult, loading: modelsLoading } = useQuery(
  AgentModelsDocument,
  () => ({ agent: props.agent.id }),
);

const sessions = computed(() => result.value?.agent.sessions ?? []);
const models = computed<AgentModelOption[]>(() => modelsResult.value?.agentModels ?? []);

const selectedId = ref<string | null>(null);
const composingNew = ref(false);

watch(sessions, (list) => {
  if (composingNew.value) return;
  if (!selectedId.value || !list.some(session => session.id === selectedId.value)) {
    selectedId.value = list[0]?.id ?? null;
  }
}, { immediate: true });

const selected = computed(() => sessions.value.find(session => session.id === selectedId.value) ?? null);

const lastChoice = computed(() => sessions.value[0] ?? null);

const newSessionModel = computed<AgentModelOption | null>(() => {
  const last = lastChoice.value?.model;
  const match = last && models.value.find(model => model.provider.id === last.provider.id && model.name === last.name);
  return match ?? models.value[0] ?? null;
});

const sessionModel = computed<AgentModelOption | null>(() => (selected.value ? selected.value.model : newSessionModel.value));
const sessionEffort = computed(() => selected.value?.effort ?? lastChoice.value?.effort ?? Effort.High);

const modelsHint = computed(() => (!modelsLoading.value && modelsResult.value && models.value.length === 0
  ? missingModelsHint(props.agent.harness)
  : null));

function startNew() {
  composingNew.value = true;
  selectedId.value = null;
}

function select(sessionId: string) {
  composingNew.value = false;
  selectedId.value = sessionId;
}

async function handleStarted(sessionId: string) {
  composingNew.value = false;
  selectedId.value = sessionId;
  await refetch();
}

function statusWord(status: AgentSessionStatus) {
  if (status === AgentSessionStatus.Idle) return null;
  const view = agentSessionStatus(status);
  return { label: view.label, tone: view.tone === 'warn' ? view.tone : 'neutral' as const };
}
</script>

<template>
  <div class="flex min-h-0 flex-1">
    <nav aria-label="Sessions" class="flex w-60 shrink-0 flex-col border-r bg-muted/30">
      <div class="shrink-0 p-3">
        <Button variant="outline" size="xs" class="w-full" @click="startNew">
          <Plus />
          New session
        </Button>
      </div>

      <div v-if="loading && !result" class="space-y-2 px-3">
        <Skeleton v-for="i in 4" :key="i" class="h-12 w-full" />
      </div>
      <div v-else-if="error && !result" class="px-3">
        <QueryError title="Couldn't load sessions" :error="error" @retry="refetch()" />
      </div>
      <p v-else-if="sessions.length === 0" class="px-4 py-2 text-xs text-muted-foreground">
        No sessions yet.
      </p>
      <ScrollArea v-else class="min-h-0 flex-1">
        <div class="divide-y border-y">
          <button
            v-for="session in sessions"
            :key="session.id"
            type="button"
            :aria-current="session.id === selectedId ? 'true' : undefined"
            class="flex w-full flex-col gap-1 px-4 py-2.5 text-left transition-colors"
            :class="session.id === selectedId ? 'bg-accent' : 'hover:bg-muted/60'"
            @click="select(session.id)"
          >
            <span class="truncate text-sm font-medium text-foreground">{{ session.title }}</span>
            <span class="flex items-center gap-2 text-xs text-muted-foreground">
              <span class="min-w-0 truncate">
                {{ memberName(session.startedBy) }} · <RelativeTime :value="session.startedAt" />
              </span>
              <span class="flex-1" />
              <Status
                v-if="statusWord(session.status)"
                variant="text"
                :tone="statusWord(session.status)!.tone"
                class="shrink-0"
              >
                {{ statusWord(session.status)!.label }}
              </Status>
            </span>
          </button>
        </div>
      </ScrollArea>
    </nav>

    <AgentSessionView
      :key="selectedId ?? 'new'"
      :agent-id="agent.id"
      :agent-name="agent.name"
      :session-id="selectedId"
      :session-model="sessionModel"
      :session-effort="sessionEffort"
      :models="models"
      :models-hint="modelsHint"
      @started="handleStarted"
      @changed="refetch()"
    />
  </div>
</template>
