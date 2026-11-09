<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { Check, Hash, MessageSquare } from '@lucide/vue';
import { graphql } from '@/gql';
import { ChatAppKind } from '@/gql/graphql';
import { chatAppKindLabel, type Agent } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import CopyButton from '@/components/CopyButton.vue';
import DialogIcon from '@/components/DialogIcon.vue';
import EmptyState from '@/components/EmptyState.vue';
import QueryError from '@/components/QueryError.vue';
import SegmentedControl from '@/components/SegmentedControl.vue';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { failureToast } from '@/components/ui/sonner';

const AddChannelChatAppsDocument = graphql(`
  query AddChannelChatApps {
    chatApps {
      id
      name
      kind
      account
    }
  }
`);

const ChatAppChannelsDocument = graphql(`
  query ChatAppChannels($id: ChatAppID!) {
    chatApp(id: $id) {
      id
      channels {
        id
        name
      }
    }
  }
`);

const AddAgentChannelDocument = graphql(`
  mutation AddAgentChannel($agent: AgentID!, $channels: [AgentChannelInput!]!) {
    setAgentChannels(agent: $agent, channels: $channels) {
      id
      channels {
        chatApp {
          id
          name
          kind
          account
          url
        }
        channel {
          id
          name
        }
      }
    }
  }
`);

const open = defineModel<boolean>('open', { default: false });

const props = defineProps<{
  agent: Agent;
}>();

const { result: appsResult, loading: appsLoading, error: appsError, refetch: refetchApps } = useQuery(AddChannelChatAppsDocument, null, () => ({ enabled: open.value }));
const apps = computed(() => appsResult.value?.chatApps ?? []);

const appId = ref('');
const selected = ref('');

const app = computed(() => apps.value.find(candidate => candidate.id === appId.value) ?? null);
const isMattermost = computed(() => app.value?.kind === ChatAppKind.Mattermost);

const appOptions = computed(() => apps.value.map(candidate => ({
  value: candidate.id,
  label: apps.value.filter(other => other.kind === candidate.kind).length > 1 ? candidate.name : chatAppKindLabel(candidate.kind),
})));

watch(open, (value) => {
  if (!value) return;
  selected.value = '';
  appId.value = apps.value[0]?.id ?? '';
});

watch(apps, (list) => {
  if (!list.some(candidate => candidate.id === appId.value)) appId.value = list[0]?.id ?? '';
});

watch(appId, () => {
  selected.value = '';
});

const { result: channelsResult, loading: channelsLoading, error: channelsError, refetch: refetchChannels } = useQuery(
  ChatAppChannelsDocument,
  () => ({ id: appId.value }),
  () => ({ enabled: open.value && isMattermost.value, fetchPolicy: 'network-only' }),
);
const channels = computed(() => channelsResult.value?.chatApp.channels ?? []);

function added(channel: string) {
  return props.agent.channels.some(existing => existing.chatApp.id === appId.value && existing.channel.id === channel);
}

const connectCommand = computed(() => `/connect@${(app.value?.account ?? '').replace(/^@/, '')}`);
const botHandle = computed(() => app.value?.account ?? 'the bot');

const { mutate: setChannels, loading: saving } = useMutation(AddAgentChannelDocument);

const canAdd = computed(() => isMattermost.value && !!selected.value && !added(selected.value) && !saving.value);

async function add() {
  if (!canAdd.value) return;
  try {
    await setChannels({
      agent: props.agent.id,
      channels: [
        ...props.agent.channels.map(existing => ({ chatApp: existing.chatApp.id, channel: existing.channel.id })),
        { chatApp: appId.value, channel: selected.value },
      ],
    });
    open.value = false;
  } catch (e: unknown) {
    failureToast('Failed to add channel', e);
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent :class="[mobileSheetClass, 'sm:max-w-md']">
      <form class="space-y-4" @submit.prevent="add">
        <DialogHeader>
          <DialogIcon :icon="MessageSquare" />
          <DialogTitle>Add channel</DialogTitle>
          <DialogDescription>Linked members of this workspace talk to {{ agent.name }} there.</DialogDescription>
        </DialogHeader>

        <div v-if="appsLoading && !appsResult" class="space-y-2">
          <Skeleton class="h-8 w-full rounded-md" />
          <Skeleton class="h-24 w-full rounded-lg" />
        </div>

        <QueryError v-else-if="appsError && !appsResult" title="Couldn't load chat apps" :error="appsError" @retry="refetchApps()" />

        <EmptyState
          v-else-if="apps.length === 0"
          size="compact"
          :icon="MessageSquare"
          title="No chat apps yet"
          description="An admin connects Telegram or Mattermost bots under Connections."
        />

        <template v-else>
          <div class="space-y-2">
            <p class="text-sm font-medium text-foreground">Chat app</p>
            <SegmentedControl v-model="appId" label="Chat app" :options="appOptions" class="flex w-full" item-class="h-7 flex-1 text-sm" />
          </div>

          <div v-if="isMattermost" class="space-y-2">
            <p class="text-sm font-medium text-foreground">Channel</p>
            <div v-if="channelsLoading && !channelsResult" class="divide-y overflow-hidden rounded-lg border">
              <div v-for="width in ['w-20', 'w-28', 'w-16']" :key="width" class="flex items-center gap-3 px-3 py-2">
                <Skeleton class="size-4 rounded-sm" />
                <Skeleton :class="['h-3', width]" />
              </div>
            </div>
            <QueryError v-else-if="channelsError && !channelsResult" title="Couldn't list the channels" :error="channelsError" @retry="refetchChannels()" />
            <p v-else-if="channels.length === 0" class="text-sm text-muted-foreground">{{ botHandle }} isn't a member of any channel yet.</p>
            <div v-else class="max-h-64 divide-y overflow-y-auto rounded-lg border">
              <button
                v-for="channel in channels"
                :key="channel.id"
                type="button"
                :disabled="added(channel.id) || saving"
                :aria-pressed="selected === channel.id"
                class="flex w-full items-center gap-3 px-3 py-2 text-left text-sm text-foreground transition-colors hover:bg-muted/50 disabled:text-muted-foreground disabled:hover:bg-transparent aria-pressed:bg-accent"
                @click="selected = channel.id"
              >
                <Hash :size="16" class="shrink-0 text-muted-foreground" />
                <span class="min-w-0 flex-1 truncate">{{ channel.name }}</span>
                <span v-if="added(channel.id)" class="shrink-0 text-xs">Added</span>
                <Check v-else-if="selected === channel.id" :size="16" class="shrink-0 text-primary" />
              </button>
            </div>
            <p class="text-xs text-muted-foreground">Only channels {{ botHandle }} is a member of. Invite it there with <span class="font-mono">/invite {{ botHandle }}</span>.</p>
          </div>

          <ol v-else class="space-y-4 pt-1">
            <li class="flex items-start gap-3">
              <span class="inline-flex size-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold">1</span>
              <p class="text-sm leading-relaxed">Add <span class="font-medium">{{ botHandle }}</span> to a group. Turn on topics there if every conversation should get its own.</p>
            </li>
            <li class="flex items-start gap-3">
              <span class="inline-flex size-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold">2</span>
              <div class="min-w-0 flex-1 space-y-2">
                <p class="text-sm leading-relaxed">Send this in the group.</p>
                <div class="flex items-center gap-2 rounded-md bg-muted/40 py-1.5 pr-1.5 pl-3">
                  <span class="min-w-0 flex-1 truncate font-mono text-sm">{{ connectCommand }}</span>
                  <CopyButton :value="connectCommand" label="Copy command" />
                </div>
              </div>
            </li>
            <li class="flex items-start gap-3">
              <span class="inline-flex size-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold">3</span>
              <p class="text-sm leading-relaxed">The bot answers with a link. It opens Lucity, where you pick the agent.</p>
            </li>
          </ol>
        </template>

        <DialogFooter>
          <template v-if="isMattermost">
            <Button type="button" variant="ghost" :disabled="saving" @click="open = false">Cancel</Button>
            <Button type="submit" :pending="saving" :disabled="!canAdd">
              {{ saving ? 'Adding…' : 'Add channel' }}
            </Button>
          </template>
          <Button v-else type="button" variant="outline" @click="open = false">Close</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
