<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Clock, Webhook, Zap } from '@lucide/vue';
import { graphql } from '@/gql';
import { AgentTriggerSource } from '@/gql/graphql';
import type { Agent, AgentTrigger } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import CodeBlock from '@/components/CodeBlock.vue';
import CredentialField from '@/components/CredentialField.vue';
import DialogIcon from '@/components/DialogIcon.vue';
import SegmentedControl from '@/components/SegmentedControl.vue';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { failureToast } from '@/components/ui/sonner';

const CreateAgentTriggerDocument = graphql(`
  mutation CreateAgentTrigger($agent: AgentID!, $input: AgentTriggerInput!) {
    createAgentTrigger(agent: $agent, input: $input) {
      token
      trigger {
        id
        name
        source
        schedule
        timeZone
        key
        instructions
        enabled
        url
        lastRunAt
      }
    }
  }
`);

const UpdateAgentTriggerDocument = graphql(`
  mutation UpdateAgentTrigger($trigger: AgentTriggerID!, $input: AgentTriggerInput!) {
    updateAgentTrigger(trigger: $trigger, input: $input) {
      id
      name
      source
      schedule
      timeZone
      key
      instructions
      enabled
      url
      lastRunAt
    }
  }
`);

const ReplaceAgentTriggerTokenDocument = graphql(`
  mutation ReplaceAgentTriggerToken($trigger: AgentTriggerID!) {
    replaceAgentTriggerToken(trigger: $trigger) {
      token
      trigger {
        id
        url
      }
    }
  }
`);


const open = defineModel<boolean>('open', { default: false });

const props = defineProps<{
  agent: Agent;
  trigger?: AgentTrigger | null;
}>();

const sources = [
  { value: AgentTriggerSource.Schedule, label: 'On a schedule', icon: Clock },
  { value: AgentTriggerSource.Webhook, label: 'From a webhook', icon: Webhook },
];

const browserTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

const name = ref('');
const source = ref<AgentTriggerSource>(AgentTriggerSource.Schedule);
const schedule = ref('');
const key = ref('');
const instructions = ref('');
const created = ref<{ url: string; token: string } | null>(null);

const editing = computed(() => !!props.trigger);
const timeZone = computed(() => props.trigger?.timeZone || browserTimeZone);
const isSchedule = computed(() => source.value === AgentTriggerSource.Schedule);

watch(open, (value) => {
  if (!value) return;
  created.value = null;
  name.value = props.trigger?.name ?? '';
  source.value = props.trigger?.source ?? AgentTriggerSource.Schedule;
  schedule.value = props.trigger?.schedule ?? '';
  key.value = props.trigger?.key ?? '';
  instructions.value = props.trigger?.instructions ?? '';
});

const { mutate: createTrigger, loading: creating } = useMutation(CreateAgentTriggerDocument, { refetchQueries: ['EnvironmentAgents'] });
const { mutate: updateTrigger, loading: updating } = useMutation(UpdateAgentTriggerDocument);
const { mutate: replaceToken, loading: replacing } = useMutation(ReplaceAgentTriggerTokenDocument);

const pending = computed(() => creating.value || updating.value || replacing.value);
const canSave = computed(() => !pending.value
  && name.value.trim() !== ''
  && instructions.value.trim() !== ''
  && (!isSchedule.value || schedule.value.trim() !== ''));

const snippet = computed(() => created.value
  ? `receivers:
  - name: ${props.agent.name}
    webhook_configs:
      - url: ${created.value.url}
        send_resolved: true
        http_config:
          authorization:
            credentials: ${created.value.token}`
  : '');

function input() {
  return {
    name: name.value.trim(),
    source: source.value,
    schedule: isSchedule.value ? schedule.value.trim() : null,
    timeZone: isSchedule.value ? timeZone.value : null,
    key: isSchedule.value ? null : key.value.trim() || null,
    instructions: instructions.value,
  };
}

async function save() {
  if (!canSave.value) return;
  try {
    if (props.trigger) {
      await updateTrigger({ trigger: props.trigger.id, input: input() });
      open.value = false;
      return;
    }

    const result = await createTrigger({ agent: props.agent.id, input: input() });
    const made = result?.data?.createAgentTrigger;

    if (made?.token && made.trigger.url) {
      created.value = { url: made.trigger.url, token: made.token };
      return;
    }

    open.value = false;
  } catch (e: unknown) {
    failureToast(props.trigger ? 'Failed to save trigger' : 'Failed to create trigger', e);
  }
}

async function newToken() {
  if (!props.trigger) return;
  try {
    const result = await replaceToken({ trigger: props.trigger.id });
    const made = result?.data?.replaceAgentTriggerToken;

    if (made?.token && made.trigger.url) created.value = { url: made.trigger.url, token: made.token };
  } catch (e: unknown) {
    failureToast('Failed to replace token', e);
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent :class="[mobileSheetClass, 'sm:max-w-lg']">
      <div v-if="created" class="space-y-4">
        <DialogHeader>
          <DialogIcon :icon="Webhook" />
          <DialogTitle>"{{ name }}" is ready</DialogTitle>
          <DialogDescription>Send events to this address with the token as a bearer token. The token is shown only now.</DialogDescription>
        </DialogHeader>

        <div class="space-y-2">
          <CredentialField label="Address" :value="created.url" />
          <CredentialField label="Token" :value="created.token" />
        </div>

        <div class="space-y-2">
          <p class="text-sm font-medium text-foreground">Alertmanager</p>
          <CodeBlock :code="snippet" :secrets="[created.token]" />
          <p class="text-xs text-muted-foreground">Route the alerts {{ agent.name }} should look at to this receiver, with <span class="font-mono">continue: true</span> to keep notifying as before.</p>
        </div>

        <DialogFooter>
          <Button type="button" @click="open = false">Done</Button>
        </DialogFooter>
      </div>

      <form v-else class="space-y-4" @submit.prevent="save">
        <DialogHeader>
          <DialogIcon :icon="Zap" />
          <DialogTitle>{{ editing ? `Trigger "${trigger?.name}"` : 'New trigger' }}</DialogTitle>
          <DialogDescription>Starts a session with {{ agent.name }} on its own<template v-if="agent.chat">, posted in {{ agent.chat.channel.name }}</template>.</DialogDescription>
        </DialogHeader>

        <fieldset :disabled="pending" class="min-w-0 space-y-4">
          <div class="space-y-2">
            <Label for="trigger-name">Name</Label>
            <Input id="trigger-name" v-model="name" placeholder="Platform alerts" autocomplete="off" maxlength="64" />
          </div>

          <div v-if="!editing" class="space-y-2">
            <p class="text-sm font-medium text-foreground">Starts</p>
            <SegmentedControl v-model="source" label="Starts" :options="sources" class="flex w-full" item-class="h-7 flex-1 text-sm" />
          </div>

          <div v-if="isSchedule" class="space-y-2">
            <Label for="trigger-schedule">Schedule</Label>
            <Input id="trigger-schedule" v-model="schedule" class="font-mono text-sm" placeholder="0 12 * * 1-5" autocomplete="off" spellcheck="false" maxlength="128" />
            <p class="text-xs text-muted-foreground">A cron expression, in {{ timeZone }}.</p>
          </div>

          <div v-else class="space-y-2">
            <Label for="trigger-key">Same session for</Label>
            <Input id="trigger-key" v-model="key" class="font-mono text-sm" placeholder="groupKey" autocomplete="off" spellcheck="false" maxlength="128" />
            <p class="text-xs text-muted-foreground">A top-level field of the JSON body. Events with the same value continue one session. Empty starts a new session for every event.</p>
          </div>

          <div v-if="editing && !isSchedule && trigger?.url" class="space-y-2">
            <CredentialField label="Address" :value="trigger.url" />
            <div class="flex items-center justify-between gap-4">
              <p class="text-xs text-muted-foreground">A new token replaces the current one, which stops working.</p>
              <Button type="button" variant="outline" size="xs" class="shrink-0" :pending="replacing" @click="newToken">New token</Button>
            </div>
          </div>

          <div class="space-y-2">
            <Label for="trigger-instructions">What {{ agent.name }} does</Label>
            <Textarea id="trigger-instructions" v-model="instructions" rows="5" maxlength="65536" placeholder="Post a short status of production: failed deploys, restarts and open alerts." />
            <p class="text-xs text-muted-foreground">The first message of every session it starts. The event follows as data.</p>
          </div>
        </fieldset>

        <DialogFooter>
          <Button type="button" variant="ghost" :disabled="pending" @click="open = false">Cancel</Button>
          <Button type="submit" :pending="creating || updating" :disabled="!canSave">
            <template v-if="editing">{{ updating ? 'Saving…' : 'Save' }}</template>
            <template v-else>{{ creating ? 'Creating…' : 'Create trigger' }}</template>
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
