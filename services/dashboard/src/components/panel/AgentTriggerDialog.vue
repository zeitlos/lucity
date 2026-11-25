<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Clock, Webhook, Zap } from '@lucide/vue';
import { graphql } from '@/gql';
import { AgentTriggerSource } from '@/gql/graphql';
import type { Agent, AgentTrigger } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import { formatDateTime } from '@/lib/format';
import { defaultSchedule, describeCron, fromCron, nextRuns, repeatOptions, toCron, weekdays, type Schedule } from '@/lib/schedule';
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
const schedule = ref<Schedule>(defaultSchedule());
const key = ref('');
const instructions = ref('');
const created = ref<{ url: string; token: string } | null>(null);

const editing = computed(() => !!props.trigger);
const timeZone = computed(() => props.trigger?.timeZone || browserTimeZone);
const isSchedule = computed(() => source.value === AgentTriggerSource.Schedule);

watch(open, (value) => {
  if (!value) return;
  created.value = null;
  example.value = 'alertmanager';
  name.value = props.trigger?.name ?? '';
  source.value = props.trigger?.source ?? AgentTriggerSource.Schedule;
  schedule.value = props.trigger?.schedule ? fromCron(props.trigger.schedule) : defaultSchedule();
  key.value = props.trigger?.key ?? '';
  instructions.value = props.trigger?.instructions ?? '';
});

const { mutate: createTrigger, loading: creating } = useMutation(CreateAgentTriggerDocument, { refetchQueries: ['EnvironmentAgents'] });
const { mutate: updateTrigger, loading: updating } = useMutation(UpdateAgentTriggerDocument);
const { mutate: replaceToken, loading: replacing } = useMutation(ReplaceAgentTriggerTokenDocument);

const cron = computed(() => toCron(schedule.value));
const runs = computed(() => nextRuns(cron.value, timeZone.value, 3));
const customReading = computed(() => (schedule.value.repeats === 'custom' && cron.value ? describeCron(cron.value) : null));

const scheduleProblem = computed(() => {
  const current = schedule.value;
  if (current.repeats === 'weekly' && current.days.length === 0) return 'Pick at least one day.';
  if (current.repeats === 'hourly' && !(current.minute >= 0 && current.minute <= 59)) return 'The minute is 0 to 59.';
  if (current.repeats === 'monthly' && !(current.dayOfMonth >= 1 && current.dayOfMonth <= 31)) return 'The day is 1 to 31.';
  if (current.repeats === 'custom' && runs.value.length === 0) return 'This isn\'t a cron expression.';
  return null;
});

function toggleDay(day: number) {
  const days = schedule.value.days;
  schedule.value.days = days.includes(day) ? days.filter(existing => existing !== day) : [...days, day];
}

const pending = computed(() => creating.value || updating.value || replacing.value);
const canSave = computed(() => !pending.value
  && name.value.trim() !== ''
  && instructions.value.trim() !== ''
  && (!isSchedule.value || !scheduleProblem.value));

const examples = [
  { value: 'alertmanager', label: 'Alertmanager' },
  { value: 'curl', label: 'curl' },
];
const example = ref('alertmanager');

const curlSnippet = computed(() => created.value
  ? `curl -X POST ${created.value.url} \\
  -H "Authorization: Bearer ${created.value.token}" \\
  -H "Content-Type: application/json" \\
  -d '${JSON.stringify({ [key.value.trim() || 'message']: 'test' })}'`
  : '');

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
    schedule: isSchedule.value ? cron.value : null,
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
      <div v-if="created" class="min-w-0 space-y-4">
        <DialogHeader>
          <DialogIcon :icon="Webhook" />
          <DialogTitle>"{{ name }}" is ready</DialogTitle>
          <DialogDescription>Send events to the address, with the token as a bearer token.</DialogDescription>
        </DialogHeader>

        <div class="space-y-2">
          <CredentialField label="Address" :value="created.url" />
          <CredentialField label="Token" :value="created.token" />
          <p class="text-xs text-muted-foreground">The token is shown only now. A lost token can be replaced, not read again.</p>
        </div>

        <div class="space-y-2">
          <CodeBlock :code="example === 'curl' ? curlSnippet : snippet" :secrets="[created.token]">
            <template #header>
              <SegmentedControl v-model="example" label="Example" :options="examples" />
            </template>
          </CodeBlock>
          <p v-if="example === 'alertmanager'" class="text-xs text-muted-foreground">Route the alerts {{ agent.name }} should look at to this receiver, with <span class="font-mono">continue: true</span> to keep notifying as before.</p>
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
            <SegmentedControl v-model="source" label="Starts" :options="sources" show-labels class="flex w-full" item-class="h-7 flex-1 text-sm" />
          </div>

          <template v-if="isSchedule">
            <div class="space-y-2">
              <p class="text-sm font-medium text-foreground">Repeats</p>
              <SegmentedControl v-model="schedule.repeats" label="Repeats" :options="repeatOptions" class="flex w-full" item-class="h-7 flex-1 text-sm" />
            </div>

            <div v-if="schedule.repeats === 'hourly'" class="space-y-2">
              <Label for="trigger-minute">Minute past the hour</Label>
              <Input id="trigger-minute" v-model.number="schedule.minute" type="number" min="0" max="59" class="w-28 tabular-nums" />
            </div>

            <div v-if="schedule.repeats === 'daily' || schedule.repeats === 'weekly' || schedule.repeats === 'monthly'" class="flex flex-wrap items-end justify-between gap-4">
              <div v-if="schedule.repeats === 'weekly'" class="space-y-2">
                <p class="text-sm font-medium text-foreground">On</p>
                <div class="flex flex-wrap gap-1.5" role="group" aria-label="Days">
                  <Button
                    v-for="day in weekdays"
                    :key="day.value"
                    type="button"
                    variant="outline"
                    size="xs"
                    class="w-10 aria-pressed:border-primary aria-pressed:bg-accent aria-pressed:text-foreground"
                    :class="schedule.days.includes(day.value) ? '' : 'text-muted-foreground'"
                    :aria-pressed="schedule.days.includes(day.value)"
                    :aria-label="day.name"
                    @click="toggleDay(day.value)"
                  >
                    {{ day.label }}
                  </Button>
                </div>
              </div>

              <div v-if="schedule.repeats === 'monthly'" class="space-y-2">
                <Label for="trigger-day">Day of the month</Label>
                <Input id="trigger-day" v-model.number="schedule.dayOfMonth" type="number" min="1" max="31" class="w-28 tabular-nums" />
              </div>

              <div class="space-y-2">
                <Label for="trigger-time">At</Label>
                <Input id="trigger-time" v-model="schedule.time" type="time" class="w-32 tabular-nums" />
              </div>
            </div>

            <div v-if="schedule.repeats === 'custom'" class="space-y-2">
              <Label for="trigger-cron">Cron expression</Label>
              <Input id="trigger-cron" v-model="schedule.cron" class="font-mono text-sm" placeholder="0 8,17 * * 1-5" autocomplete="off" spellcheck="false" maxlength="128" />
              <p v-if="customReading" class="text-xs text-muted-foreground">{{ customReading }}.</p>
            </div>

            <p v-if="scheduleProblem && (schedule.repeats !== 'custom' || schedule.cron.trim() !== '')" class="text-xs text-destructive">{{ scheduleProblem }}</p>
            <p v-else-if="runs.length > 0" class="text-xs text-muted-foreground tabular-nums">{{ timeZone }} time. Next: {{ runs.map(run => formatDateTime(run)).join(' · ') }}</p>
          </template>

          <div class="space-y-2">
            <Label for="trigger-instructions">What {{ agent.name }} does</Label>
            <Textarea id="trigger-instructions" v-model="instructions" rows="3" maxlength="65536" placeholder="Post a short status of production: failed deploys, restarts and open alerts." />
            <p class="text-xs text-muted-foreground">The first message of every session it starts. The event follows as data.</p>
          </div>

          <div v-if="!isSchedule" class="space-y-2">
            <Label for="trigger-key">Group events by</Label>
            <Input id="trigger-key" v-model="key" class="font-mono text-sm" placeholder="groupKey" autocomplete="off" spellcheck="false" maxlength="128" />
            <p class="text-xs text-muted-foreground">A top-level field of the JSON body. Events with the same value continue one session. Leave it empty to start a session for every event.</p>
          </div>

          <div v-if="editing && !isSchedule && trigger?.url" class="space-y-2">
            <CredentialField label="Address" :value="trigger.url" />
            <div class="flex items-center justify-between gap-4">
              <p class="text-xs text-muted-foreground">A new token replaces the current one, which stops working.</p>
              <Button type="button" variant="outline" size="xs" class="shrink-0" :pending="replacing" @click="newToken">New token</Button>
            </div>
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
