<script setup lang="ts">
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { MessageSquare, Plus, Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import { ChatAppKind } from '@/gql/graphql';
import { chatAppKindLabel, type Agent, type AgentChat } from '@/lib/agents';
import BrandTile from '@/components/BrandTile.vue';
import EmptyState from '@/components/EmptyState.vue';
import IconButton from '@/components/IconButton.vue';
import { Button } from '@/components/ui/button';
import { failureToast } from '@/components/ui/sonner';
import AddChannelDialog from './AddChannelDialog.vue';

const RemoveAgentChatDocument = graphql(`
  mutation RemoveAgentChat($agent: AgentID!) {
    removeAgentChat(agent: $agent) {
      id
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
`);

const props = defineProps<{
  agent: Agent;
}>();

const addOpen = ref(false);

const { mutate: removeChat, loading: removing } = useMutation(RemoveAgentChatDocument);

function where(chat: AgentChat) {
  if (chat.chatApp.kind === ChatAppKind.Mattermost && chat.chatApp.url) {
    try {
      return `Mattermost · ${new URL(chat.chatApp.url).host}`;
    } catch {
      return 'Mattermost';
    }
  }

  return `${chatAppKindLabel(chat.chatApp.kind)} · ${chat.chatApp.account}`;
}

async function remove() {
  try {
    await removeChat({ agent: props.agent.id });
  } catch (e: unknown) {
    failureToast('Failed to remove channel', e);
  }
}
</script>

<template>
  <section class="space-y-4">
    <div class="flex items-start justify-between gap-4">
      <p class="text-sm text-muted-foreground">
        Linked members of this workspace talk to {{ agent.name }} in this channel and answer its approvals there. Sessions its triggers start are posted there too.
      </p>
      <Button v-if="!agent.chat" variant="outline" size="xs" class="shrink-0" @click="addOpen = true">
        <Plus />
        Add channel
      </Button>
    </div>

    <EmptyState
      v-if="!agent.chat"
      size="compact"
      :icon="MessageSquare"
      title="No channel yet"
      description="Add a Telegram chat or a Mattermost channel."
    />
    <div v-else class="-mx-4 divide-y border-y">
      <div class="flex items-center gap-3 py-2.5 pr-3 pl-4">
        <BrandTile :name="chatAppKindLabel(agent.chat.chatApp.kind)" />
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm font-medium text-foreground">{{ agent.chat.channel.name }}</p>
          <p class="truncate text-xs text-muted-foreground">{{ where(agent.chat) }}</p>
        </div>
        <Button variant="outline" size="xs" class="shrink-0" :disabled="removing" @click="addOpen = true">Change</Button>
        <IconButton
          :label="`Remove ${agent.chat.channel.name}`"
          size="icon-xs"
          class="shrink-0"
          :pending="removing"
          @click="remove"
        >
          <Trash2 />
        </IconButton>
      </div>
    </div>

    <AddChannelDialog v-model:open="addOpen" :agent="agent" />
  </section>
</template>
