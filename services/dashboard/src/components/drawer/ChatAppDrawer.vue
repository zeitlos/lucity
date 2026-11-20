<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { ArrowLeft, RotateCw, Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import { CredentialStatus } from '@/gql/graphql';
import { chatAppKindLabel } from '@/lib/agents';
import { credentialStatus } from '@/lib/status';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import IconButton from '@/components/IconButton.vue';
import QueryError from '@/components/QueryError.vue';
import RelativeTime from '@/components/RelativeTime.vue';
import ReplaceSecretDialog from '@/components/ReplaceSecretDialog.vue';
import ConnectionAgents from '@/components/drawer/ConnectionAgents.vue';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Status } from '@/components/ui/status';
import { failureToast, successToast } from '@/components/ui/sonner';

const ChatAppDocument = graphql(`
  query ChatApp($id: ChatAppID!) {
    chatApp(id: $id) {
      id
      name
      kind
      account
      url
      credentialHint
      status
      checkedAt
      agents {
        id
        name
        harness
      }
    }
  }
`);

const ReplaceChatAppTokenDocument = graphql(`
  mutation ReplaceChatAppToken($chatApp: ChatAppID!, $token: String!) {
    replaceChatAppToken(chatApp: $chatApp, token: $token) {
      id
      account
      credentialHint
      status
      checkedAt
    }
  }
`);

const RefreshChatAppDocument = graphql(`
  mutation RefreshChatApp($chatApp: ChatAppID!) {
    refreshChatApp(chatApp: $chatApp) {
      id
      account
      status
      checkedAt
    }
  }
`);

const DeleteChatAppDocument = graphql(`
  mutation DeleteChatApp($chatApp: ChatAppID!) {
    deleteChatApp(chatApp: $chatApp)
  }
`);

const props = defineProps<{
  workspace: string;
  name: string;
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const router = useRouter();
const id = computed(() => `${props.workspace}/${props.name}`);

const { result, loading, error, refetch } = useQuery(ChatAppDocument, () => ({ id: id.value }));
const app = computed(() => result.value?.chatApp ?? null);
const status = computed(() => (app.value ? credentialStatus(app.value.status, 'Token') : null));

const { mutate: replaceToken, loading: replacing } = useMutation(ReplaceChatAppTokenDocument);
const { mutate: refresh, loading: refreshing } = useMutation(RefreshChatAppDocument);
const { mutate: deleteChatApp, loading: deleting } = useMutation(DeleteChatAppDocument);

const replaceOpen = ref(false);

async function handleReplace(token: string) {
  try {
    await replaceToken({ chatApp: id.value, token });
    replaceOpen.value = false;
  } catch (e: unknown) {
    failureToast('Failed to replace token', e);
  }
}

async function handleRefresh() {
  try {
    await refresh({ chatApp: id.value });
  } catch (e: unknown) {
    failureToast('Failed to check chat app', e);
  }
}

async function handleDelete() {
  try {
    await deleteChatApp({ chatApp: id.value });
    successToast(`Chat app "${props.name}" deleted`);
    router.push({ name: 'workspace', params: { drawer: 'connections' } });
  } catch (e: unknown) {
    failureToast('Failed to delete chat app', e);
  }
}
</script>

<template>
  <ContextDrawer :title="app?.name ?? name" @close="emit('close')">
    <template #leading>
      <IconButton label="Back to connections" class="-ml-1 shrink-0" as-child>
        <RouterLink :to="{ name: 'workspace', params: { drawer: 'connections' } }">
          <ArrowLeft />
        </RouterLink>
      </IconButton>
    </template>

    <template v-if="isAdmin && app" #actions>
      <Button variant="outline" size="xs" :pending="refreshing" @click="handleRefresh">
        <RotateCw />
        {{ refreshing ? 'Checking…' : 'Check again' }}
      </Button>
    </template>

    <div v-if="loading && !result" class="space-y-8">
      <div class="flex h-5 items-center">
        <Skeleton class="h-3 w-56 max-w-full" />
      </div>

      <section class="space-y-2">
        <h3 class="text-sm font-medium text-muted-foreground">Token</h3>
        <div class="flex h-11 items-center rounded-md bg-muted/40 pr-2 pl-3">
          <Skeleton class="h-3 w-32" />
        </div>
      </section>
    </div>

    <QueryError
      v-else-if="error && !result"
      title="Couldn't load this chat app"
      :error="error"
      @retry="refetch()"
    />

    <div v-else-if="app" class="space-y-8">
      <p class="text-sm text-muted-foreground">
        {{ chatAppKindLabel(app.kind) }} bot {{ app.account }}<template v-if="app.url"> on <span class="font-mono break-all">{{ app.url }}</span></template>
      </p>

      <section class="space-y-2">
        <h3 class="text-sm font-medium text-muted-foreground">Token</h3>
        <div class="flex items-center gap-2 rounded-md bg-muted/40 py-2 pr-2 pl-3">
          <span class="min-w-0 flex-1 truncate font-mono text-xs text-foreground">••••••••••••{{ app.credentialHint }}</span>
          <Status
            v-if="status && app.status !== CredentialStatus.Valid"
            variant="text"
            :tone="status.tone"
            class="shrink-0 text-xs"
          >
            {{ status.label }}
          </Status>
          <Button v-if="isAdmin" variant="outline" size="xs" class="shrink-0" @click="replaceOpen = true">Replace</Button>
        </div>
        <p class="text-xs text-muted-foreground">
          <template v-if="app.checkedAt">Checked <RelativeTime :value="app.checkedAt" />. </template>
          Agents never see it.
        </p>
      </section>

      <section class="space-y-3">
        <div class="space-y-1 text-sm text-muted-foreground/80">
          <h3 class="text-sm font-medium text-muted-foreground">Agents</h3>
          <p>Agents with a channel in this chat app. Add one from an agent's Channels tab.</p>
        </div>
        <ConnectionAgents :agents="app.agents" empty="No agent talks through this chat app yet." />
      </section>

      <div v-if="isAdmin" class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3">
        <div>
          <p class="text-sm font-medium text-foreground">Delete chat app</p>
          <p class="text-xs text-muted-foreground">Its token is deleted, and every agent leaves its chats.</p>
        </div>
        <ConfirmDialog
          :title="`Delete chat app &quot;${app.name}&quot;?`"
          description="Its token is deleted, and every agent's channels in it are removed. This can't be undone."
          confirm-label="Delete chat app"
          pending-label="Deleting…"
          :pending="deleting"
          destructive
          @confirm="handleDelete"
        >
          <template #trigger>
            <Button variant="destructive">
              <Trash2 />
              Delete
            </Button>
          </template>
        </ConfirmDialog>
      </div>
    </div>

    <ReplaceSecretDialog
      v-model:open="replaceOpen"
      :title="`Replace the token of &quot;${app?.name ?? name}&quot;`"
      description="The token is checked by asking the chat app who the bot is. The old one stays until the check passes."
      label="Bot token"
      confirm-label="Replace token"
      :pending="replacing"
      @submit="handleReplace"
    />
  </ContextDrawer>
</template>
