<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { Bot, MessageSquare } from '@lucide/vue';
import { graphql } from '@/gql';
import { chatAppKindLabel } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import BrandTile from '@/components/BrandTile.vue';
import ChoiceRows from '@/components/ChoiceRows.vue';
import DialogIcon from '@/components/DialogIcon.vue';
import QueryError from '@/components/QueryError.vue';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
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
          chat {
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

const SetAgentChatFromLinkDocument = graphql(`
  mutation SetAgentChatFromLink($agent: AgentID!, $chatApp: ChatAppID!, $channel: String!) {
    setAgentChat(agent: $agent, chatApp: $chatApp, channel: $channel) {
      id
      chat {
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
const agentOptions = computed(() => groups.value.flatMap(group => group.agents.map(agent => ({ value: agent.id, label: agent.name, hint: group.label }))));
const agentId = ref('');

watch(agents, (list) => {
  if (!agentId.value && list.length === 1) agentId.value = list[0]!.id;
}, { immediate: true });

const owner = computed(() => agents.value.find(agent => agent.chat?.chatApp.id === chatAppId.value && agent.chat?.channel.id === props.channel) ?? null);

watch(open, (value) => {
  if (!value) emit('close');
});

const { mutate: setChat, loading: saving } = useMutation(SetAgentChatFromLinkDocument);

const canAdd = computed(() => !saving.value && !!agentId.value && !!found.value && !owner.value);

async function add() {
  const agent = agents.value.find(candidate => candidate.id === agentId.value);
  if (!canAdd.value || !agent || !found.value) return;
  try {
    await setChat({ agent: agent.id, chatApp: chatAppId.value, channel: found.value.id });
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
          <DialogTitle>Add channel</DialogTitle>
          <DialogDescription v-if="app && found">
            Pick the agent the {{ chatAppKindLabel(app.kind) }} chat "{{ found.name }}" talks to. Linked members there reach it.
          </DialogDescription>
          <DialogDescription v-else>Pick the agent this chat talks to.</DialogDescription>
        </DialogHeader>

        <div v-if="loading && !result" class="space-y-1">
          <div v-for="width in ['w-24', 'w-32']" :key="width" class="flex items-center gap-2.5 p-1">
            <Skeleton class="size-8 rounded-sm" />
            <Skeleton :class="['h-3', width]" />
          </div>
        </div>

        <QueryError v-else-if="error && !result" title="Couldn't look up this chat" :error="error" @retry="refetch()" />

        <template v-else-if="app && found">
          <p v-if="owner" class="text-sm text-muted-foreground">This chat already talks to "{{ owner.name }}". Remove it there first.</p>

          <p v-else-if="agents.length === 0" class="text-sm text-muted-foreground">This workspace has no agent yet. Create one on an environment's canvas first.</p>

          <div v-else class="space-y-2">
            <p id="add-channel-agent" class="text-sm font-medium text-foreground">Agent</p>
            <ChoiceRows v-model="agentId" :options="agentOptions" :columns="1" aria-labelledby="add-channel-agent" class="max-h-64 overflow-y-auto">
              <template #visual="{ option }">
                <BrandTile :name="option.label" :icon="Bot" />
              </template>
            </ChoiceRows>
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
