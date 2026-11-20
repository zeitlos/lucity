<script setup lang="ts">
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { MessageSquare, Plus, Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import { ChatAppKind } from '@/gql/graphql';
import { chatAppKindLabel, type Agent, type AgentChannel } from '@/lib/agents';
import BrandTile from '@/components/BrandTile.vue';
import EmptyState from '@/components/EmptyState.vue';
import IconButton from '@/components/IconButton.vue';
import { Button } from '@/components/ui/button';
import { failureToast } from '@/components/ui/sonner';
import AddChannelDialog from './AddChannelDialog.vue';

const SetAgentChannelsDocument = graphql(`
  mutation SetAgentChannels($agent: AgentID!, $channels: [AgentChannelInput!]!) {
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
  agent: Agent;
}>();

const addOpen = ref(false);
const removing = ref<string | null>(null);

const { mutate: setChannels, loading: saving } = useMutation(SetAgentChannelsDocument);

function key(channel: AgentChannel) {
  return `${channel.chatApp.id}\n${channel.channel.id}`;
}

function where(channel: AgentChannel) {
  if (channel.chatApp.kind === ChatAppKind.Mattermost && channel.chatApp.url) {
    try {
      return `Mattermost · ${new URL(channel.chatApp.url).host}`;
    } catch {
      return 'Mattermost';
    }
  }

  return `${chatAppKindLabel(channel.chatApp.kind)} · ${channel.chatApp.account}`;
}

async function remove(channel: AgentChannel) {
  removing.value = key(channel);
  try {
    await setChannels({
      agent: props.agent.id,
      channels: props.agent.channels
        .filter(existing => key(existing) !== key(channel))
        .map(existing => ({ chatApp: existing.chatApp.id, channel: existing.channel.id })),
    });
  } catch (e: unknown) {
    failureToast('Failed to remove channel', e);
  } finally {
    removing.value = null;
  }
}
</script>

<template>
  <section class="space-y-4">
    <div class="flex items-start justify-between gap-4">
      <p class="text-sm text-muted-foreground">
        Linked members of this workspace talk to {{ agent.name }} in these channels and answer its approvals there.
      </p>
      <Button variant="outline" size="xs" class="shrink-0" @click="addOpen = true">
        <Plus />
        Add channel
      </Button>
    </div>

    <EmptyState
      v-if="agent.channels.length === 0"
      size="compact"
      :icon="MessageSquare"
      title="No channels yet"
      description="Add a Telegram chat or a Mattermost channel."
    />
    <div v-else class="-mx-4 divide-y border-y">
      <div v-for="channel in agent.channels" :key="key(channel)" class="flex items-center gap-3 py-2.5 pr-3 pl-4">
        <BrandTile :name="chatAppKindLabel(channel.chatApp.kind)" />
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm font-medium text-foreground">{{ channel.channel.name }}</p>
          <p class="truncate text-xs text-muted-foreground">{{ where(channel) }}</p>
        </div>
        <IconButton
          :label="`Remove ${channel.channel.name}`"
          size="icon-xs"
          class="shrink-0"
          :pending="removing === key(channel)"
          :disabled="saving"
          @click="remove(channel)"
        >
          <Trash2 />
        </IconButton>
      </div>
    </div>

    <AddChannelDialog v-model:open="addOpen" :agent="agent" />
  </section>
</template>
