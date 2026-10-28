<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import { useMutation, useQuery, useSubscription } from '@vue/apollo-composable';
import { Bot, MessagesSquare } from '@lucide/vue';
import { graphql } from '@/gql';
import { AgentSessionStatus, AgentTurnEndReason, type AgentApprovalDecision, type Effort } from '@/gql/graphql';
import { buildTranscript, memberName, sessionActivity, type SessionEvent } from '@/lib/agentSession';
import type { AgentModelOption } from '@/lib/agents';
import { formatTime } from '@/lib/format';
import EmptyState from '@/components/EmptyState.vue';
import QueryError from '@/components/QueryError.vue';
import { Skeleton } from '@/components/ui/skeleton';
import { failureToast } from '@/components/ui/sonner';
import AgentApprovalCard from './AgentApprovalCard.vue';
import AgentComposer from './AgentComposer.vue';
import AgentToolCalls from './AgentToolCalls.vue';

const AgentSessionDocument = graphql(`
  query AgentSession($session: AgentSessionID!) {
    agentSession(id: $session) {
      id
      events {
        __typename
        ... on AgentUserMessage { id time text author { id name email } }
        ... on AgentMessage { id time text }
        ... on AgentToolCall { id time callId connector tool title status input output }
        ... on AgentApproval { id time approval toolCallId: callId title approvalStatus: status decidedBy { id name email } }
        ... on AgentTurnEnd { id time reason error }
      }
    }
  }
`);

const AgentSessionEventsDocument = graphql(`
  subscription AgentSessionEvents($session: AgentSessionID!, $after: ID) {
    agentSessionEvents(session: $session, after: $after) {
      __typename
      ... on AgentUserMessage { id time text author { id name email } }
      ... on AgentMessage { id time text }
      ... on AgentToolCall { id time callId connector tool title status input output }
      ... on AgentApproval { id time approval toolCallId: callId title approvalStatus: status decidedBy { id name email } }
      ... on AgentTurnEnd { id time reason error }
    }
  }
`);

const StartAgentSessionDocument = graphql(`
  mutation StartAgentSession($agent: AgentID!, $message: String!, $model: AgentModelInput, $effort: Effort) {
    startAgentSession(agent: $agent, message: $message, model: $model, effort: $effort) {
      id
      title
      startedAt
      status
      effort
      startedBy { id name email }
      model { name provider { id name } }
    }
  }
`);

const SendAgentMessageDocument = graphql(`
  mutation SendAgentMessage($session: AgentSessionID!, $message: String!) {
    sendAgentMessage(session: $session, message: $message) {
      id
      status
    }
  }
`);

const SetAgentSessionModelDocument = graphql(`
  mutation SetAgentSessionModel($session: AgentSessionID!, $model: AgentModelInput!, $effort: Effort!) {
    setAgentSessionModel(session: $session, model: $model, effort: $effort) {
      id
      effort
      model { name provider { id name } }
    }
  }
`);

const AnswerAgentApprovalDocument = graphql(`
  mutation AnswerAgentApproval($session: AgentSessionID!, $approval: ID!, $decision: AgentApprovalDecision!) {
    answerAgentApproval(session: $session, approval: $approval, decision: $decision) {
      id
      status
    }
  }
`);

const StopAgentSessionDocument = graphql(`
  mutation StopAgentSession($session: AgentSessionID!) {
    stopAgentSession(session: $session) {
      id
      status
    }
  }
`);

const props = defineProps<{
  agentId: string;
  agentName: string;
  sessionId: string | null;
  sessionModel: AgentModelOption | null;
  sessionEffort: Effort;
  models: AgentModelOption[];
  modelsHint: string | null;
}>();

const emit = defineEmits<{
  (e: 'started', sessionId: string): void;
  (e: 'changed'): void;
}>();

const events = ref<SessionEvent[]>([]);
const after = ref<string | null>(null);
const live = ref(false);
const liveFailed = ref(false);

const { result, loading, error, refetch, onResult } = useQuery(
  AgentSessionDocument,
  () => ({ session: props.sessionId ?? '' }),
  () => ({
    enabled: !!props.sessionId,
    fetchPolicy: 'no-cache',
    pollInterval: liveFailed.value ? 3000 : 0,
  }),
);

onResult(({ data }) => {
  if (!data?.agentSession) return;
  events.value = data.agentSession.events;
  if (!live.value) {
    after.value = data.agentSession.events.at(-1)?.id ?? null;
    live.value = true;
  }
});

const { onResult: onEvent, onError: onLiveError } = useSubscription(
  AgentSessionEventsDocument,
  () => ({ session: props.sessionId ?? '', after: after.value }),
  () => ({
    enabled: !!props.sessionId && live.value && !liveFailed.value,
    fetchPolicy: 'no-cache',
  }),
);

onEvent(({ data }) => {
  const event = data?.agentSessionEvents;
  if (!event || events.value.some(existing => existing.id === event.id)) return;
  events.value = [...events.value, event];
});

onLiveError(() => {
  liveFailed.value = true;
});

const transcript = computed(() => buildTranscript(events.value));
const activity = computed(() => sessionActivity(events.value));
const working = computed(() => activity.value !== AgentSessionStatus.Idle);

watch(
  () => events.value.at(-1)?.__typename,
  (type, previous) => {
    if (type === 'AgentTurnEnd' && previous && previous !== 'AgentTurnEnd') emit('changed');
  },
);

const scroller = ref<HTMLElement | null>(null);
const stickToBottom = ref(true);

function onScroll() {
  const el = scroller.value;
  if (!el) return;
  stickToBottom.value = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
}

watch(
  () => [events.value.length, events.value.at(-1)?.id],
  () => {
    if (!stickToBottom.value) return;
    nextTick(() => {
      if (scroller.value) scroller.value.scrollTop = scroller.value.scrollHeight;
    });
  },
);

const draftModel = ref<AgentModelOption | null>(props.sessionModel);
const draftEffort = ref<Effort>(props.sessionEffort);

watch(() => [props.sessionModel, props.sessionEffort] as const, ([model, effort]) => {
  draftModel.value = model;
  draftEffort.value = effort;
});

const { mutate: startSession, loading: starting } = useMutation(StartAgentSessionDocument);
const { mutate: sendMessage, loading: sending } = useMutation(SendAgentMessageDocument);
const { mutate: setSessionModel } = useMutation(SetAgentSessionModelDocument);
const { mutate: answerApproval } = useMutation(AnswerAgentApprovalDocument);
const { mutate: stopSession, loading: stopping } = useMutation(StopAgentSessionDocument);

const composer = ref<InstanceType<typeof AgentComposer> | null>(null);

function modelInput(model: AgentModelOption) {
  return { provider: model.provider.id, name: model.name };
}

async function send(message: string) {
  try {
    if (props.sessionId) {
      await sendMessage({ session: props.sessionId, message });
    } else {
      const model = draftModel.value;
      const res = await startSession({
        agent: props.agentId,
        message,
        model: model ? modelInput(model) : null,
        effort: draftEffort.value,
      });
      const id = res?.data?.startAgentSession.id;
      if (id) emit('started', id);
    }
    stickToBottom.value = true;
    composer.value?.clear();
  } catch (e: unknown) {
    failureToast(props.sessionId ? 'Failed to send message' : 'Failed to start session', e);
  }
}

async function changeModel(model: AgentModelOption, effort: Effort) {
  const previous = { model: draftModel.value, effort: draftEffort.value };
  draftModel.value = model;
  draftEffort.value = effort;
  if (!props.sessionId) return;
  try {
    await setSessionModel({ session: props.sessionId, model: modelInput(model), effort });
  } catch (e: unknown) {
    draftModel.value = previous.model;
    draftEffort.value = previous.effort;
    failureToast('Failed to change model', e);
  }
}

async function stop() {
  if (!props.sessionId) return;
  try {
    await stopSession({ session: props.sessionId });
  } catch (e: unknown) {
    failureToast('Failed to stop session', e);
  }
}

const answering = ref<{ approval: string; decision: AgentApprovalDecision } | null>(null);

async function answer(approval: string, decision: AgentApprovalDecision) {
  if (!props.sessionId) return;
  answering.value = { approval, decision };
  try {
    await answerApproval({ session: props.sessionId, approval, decision });
    emit('changed');
  } catch (e: unknown) {
    failureToast('Failed to answer approval', e);
  } finally {
    answering.value = null;
  }
}

function retryLive() {
  liveFailed.value = false;
  live.value = false;
  refetch();
}
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div ref="scroller" class="min-h-0 flex-1 overflow-y-auto px-4 py-4" @scroll="onScroll">
      <EmptyState
        v-if="!sessionId"
        size="compact"
        :icon="MessagesSquare"
        title="New session"
        :description="`Ask ${agentName} something. Everyone in the workspace can see the session.`"
      />

      <div v-else-if="loading && !result" class="space-y-4">
        <Skeleton class="ml-auto h-12 w-2/3" />
        <Skeleton class="h-20 w-5/6" />
        <Skeleton class="h-24 w-5/6" />
      </div>

      <QueryError
        v-else-if="error && !result"
        title="Couldn't load this session"
        :error="error"
        @retry="refetch()"
      />

      <div v-else class="space-y-5">
        <template v-for="block in transcript" :key="block.key">
          <div v-if="block.kind === 'user'" class="flex flex-col items-end gap-1">
            <span class="text-xs text-muted-foreground">
              {{ memberName(block.event.author) }} · {{ formatTime(block.event.time) }}
            </span>
            <div class="max-w-[85%] rounded-lg bg-muted px-3 py-2 text-sm whitespace-pre-wrap break-words text-foreground">
              {{ block.event.text }}
            </div>
          </div>

          <div v-else class="space-y-2.5">
            <span class="flex items-center gap-1.5 text-xs text-muted-foreground">
              <Bot :size="14" class="shrink-0" />
              {{ agentName }} · {{ formatTime(block.time) }}
            </span>
            <template v-for="item in block.items" :key="item.key">
              <div
                v-if="item.kind === 'message'"
                class="text-sm leading-relaxed whitespace-pre-wrap break-words text-foreground"
              >
                {{ item.event.text }}
              </div>
              <AgentToolCalls v-else-if="item.kind === 'tools'" :calls="item.calls" />
              <AgentApprovalCard
                v-else-if="item.kind === 'approval'"
                :approval="item.event"
                :call="item.call"
                :agent-name="agentName"
                :answering="answering?.approval === item.event.approval ? answering.decision : null"
                @answer="answer(item.event.approval, $event)"
              />
              <p
                v-else-if="item.event.reason === AgentTurnEndReason.Failed"
                class="text-xs text-status-danger"
              >
                The turn failed{{ item.event.error ? `: ${item.event.error}` : '.' }}
              </p>
              <p v-else class="text-xs text-muted-foreground">Stopped.</p>
            </template>
          </div>
        </template>
      </div>
    </div>

    <div class="shrink-0 space-y-2 px-4 pb-4">
      <p v-if="liveFailed" class="text-xs text-muted-foreground">
        Live updates paused, refreshing every few seconds.
        <button type="button" class="text-foreground underline underline-offset-2" @click="retryLive">Reconnect</button>
      </p>
      <p v-if="modelsHint" class="text-xs text-muted-foreground">
        {{ modelsHint }}
        <RouterLink
          :to="{ name: 'workspace', params: { drawer: 'connections' } }"
          class="text-foreground underline underline-offset-2"
        >Connect a model provider</RouterLink>
      </p>
      <AgentComposer
        ref="composer"
        :agent-name="agentName"
        :models="models"
        :model="draftModel"
        :effort="draftEffort"
        :working="working"
        :sending="sending || starting"
        :stopping="stopping"
        :disabled="models.length === 0"
        @send="send"
        @stop="stop"
        @change="changeModel"
      />
    </div>
  </div>
</template>
