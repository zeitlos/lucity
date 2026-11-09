<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { MessageSquare } from '@lucide/vue';
import { graphql } from '@/gql';
import { chatAppKindLabel } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import BrandTile from '@/components/BrandTile.vue';
import DialogIcon from '@/components/DialogIcon.vue';
import QueryError from '@/components/QueryError.vue';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { failureToast, successToast } from '@/components/ui/sonner';

const AddChannelLinkDocument = graphql(`
  query AddChannelLink($chatApp: ChatAppID!, $channel: String!) {
    chatApp(id: $chatApp) {
      id
      name
      kind
      account
    }
    chatChannel(chatApp: $chatApp, channel: $channel) {
      id
      name
    }
    projects {
      id
      name
      environments {
        id
        name
        agents {
          id
          name
          channels {
            chatApp {
              id
            }
            channel {
              id
            }
          }
        }
      }
    }
  }
`);

const SetAgentChannelsFromLinkDocument = graphql(`
  mutation SetAgentChannelsFromLink($agent: AgentID!, $channels: [AgentChannelInput!]!) {
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

const props = defineProps<{
  workspace: string;
  chatApp: string;
  channel: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const open = ref(true);
const chatAppId = computed(() => `${props.workspace}/${props.chatApp}`);

const { result, loading, error, refetch } = useQuery(AddChannelLinkDocument, () => ({ chatApp: chatAppId.value, channel: props.channel }));

const app = computed(() => result.value?.chatApp ?? null);
const found = computed(() => result.value?.chatChannel ?? null);

const groups = computed(() => (result.value?.projects ?? []).flatMap(project => project.environments
  .filter(environment => environment.agents.length > 0)
  .map(environment => ({
    id: environment.id,
    label: `${project.name} / ${environment.name}`,
    agents: environment.agents,
  }))));

const agents = computed(() => groups.value.flatMap(group => group.agents));
const agentId = ref('');

watch(agents, (list) => {
  if (!agentId.value && list.length === 1) agentId.value = list[0]!.id;
}, { immediate: true });

const owner = computed(() => agents.value.find(agent => agent.channels.some(channel => channel.chatApp.id === chatAppId.value && channel.channel.id === props.channel)) ?? null);

watch(open, (value) => {
  if (!value) emit('close');
});

const { mutate: setChannels, loading: saving } = useMutation(SetAgentChannelsFromLinkDocument);

const canAdd = computed(() => !saving.value && !!agentId.value && !!found.value && !owner.value);

async function add() {
  const agent = agents.value.find(candidate => candidate.id === agentId.value);
  if (!canAdd.value || !agent || !found.value) return;
  try {
    await setChannels({
      agent: agent.id,
      channels: [
        ...agent.channels.map(channel => ({ chatApp: channel.chatApp.id, channel: channel.channel.id })),
        { chatApp: chatAppId.value, channel: found.value.id },
      ],
    });
    successToast(`"${found.value.name}" added to "${agent.name}"`);
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
          <DialogTitle>Add a channel</DialogTitle>
          <DialogDescription v-if="app && found">
            The {{ chatAppKindLabel(app.kind) }} chat "{{ found.name }}" asked to talk to an agent of this workspace. Linked members there reach the agent you pick.
          </DialogDescription>
          <DialogDescription v-else>Pick the agent this chat talks to.</DialogDescription>
        </DialogHeader>

        <div v-if="loading && !result" class="flex items-center gap-3 rounded-lg border p-3">
          <Skeleton class="size-8 rounded-sm" />
          <div class="flex-1 space-y-1.5">
            <Skeleton class="h-3 w-28" />
            <Skeleton class="h-2.5 w-44" />
          </div>
        </div>

        <QueryError v-else-if="error && !result" title="Couldn't look up this chat" :error="error" @retry="refetch()" />

        <template v-else-if="app && found">
          <div class="flex items-center gap-3 rounded-lg border p-3">
            <BrandTile :name="chatAppKindLabel(app.kind)" />
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium text-foreground">{{ found.name }}</p>
              <p class="truncate text-xs text-muted-foreground">{{ chatAppKindLabel(app.kind) }} · {{ app.account }} is a member</p>
            </div>
          </div>

          <p v-if="owner" class="text-sm text-muted-foreground">This chat already talks to "{{ owner.name }}". Remove it there first.</p>

          <div v-else class="space-y-2">
            <Label for="add-channel-agent">Agent</Label>
            <Select v-model="agentId" :disabled="saving || agents.length === 0">
              <SelectTrigger id="add-channel-agent" class="w-full">
                <SelectValue placeholder="Pick an agent" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup v-for="group in groups" :key="group.id">
                  <SelectLabel>{{ group.label }}</SelectLabel>
                  <SelectItem v-for="agent in group.agents" :key="agent.id" :value="agent.id">{{ agent.name }}</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <p v-if="agents.length === 0" class="text-xs text-muted-foreground">This workspace has no agent yet. Create one on an environment's canvas first.</p>
          </div>
        </template>

        <DialogFooter>
          <Button type="button" variant="ghost" :disabled="saving" @click="open = false">Cancel</Button>
          <Button type="submit" :pending="saving" :disabled="!canAdd">
            {{ saving ? 'Adding…' : 'Add channel' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
