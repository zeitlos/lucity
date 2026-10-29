<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { ArrowLeft, RotateCw, Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import { CredentialStatus } from '@/gql/graphql';
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

const ConnectorDocument = graphql(`
  query Connector($id: ConnectorID!) {
    connector(id: $id) {
      id
      name
      url
      credentialHint
      status
      checkedAt
      tools {
        name
        description
      }
      agents {
        id
        name
        harness
      }
    }
  }
`);

const ReplaceConnectorCredentialDocument = graphql(`
  mutation ReplaceConnectorCredential($connector: ConnectorID!, $credential: String!) {
    replaceConnectorCredential(connector: $connector, credential: $credential) {
      id
      credentialHint
      status
      checkedAt
      tools {
        name
        description
      }
    }
  }
`);

const RefreshConnectorDocument = graphql(`
  mutation RefreshConnector($connector: ConnectorID!) {
    refreshConnector(connector: $connector) {
      id
      status
      checkedAt
      tools {
        name
        description
      }
    }
  }
`);

const DeleteConnectorDocument = graphql(`
  mutation DeleteConnector($connector: ConnectorID!) {
    deleteConnector(connector: $connector)
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

const { result, loading, error, refetch } = useQuery(ConnectorDocument, () => ({ id: id.value }));
const connector = computed(() => result.value?.connector ?? null);
const status = computed(() => (connector.value ? credentialStatus(connector.value.status, 'Token') : null));

const { mutate: replaceCredential, loading: replacing } = useMutation(ReplaceConnectorCredentialDocument);
const { mutate: refresh, loading: refreshing } = useMutation(RefreshConnectorDocument);
const { mutate: deleteConnector, loading: deleting } = useMutation(DeleteConnectorDocument);

const replaceOpen = ref(false);

async function handleReplace(credential: string) {
  try {
    await replaceCredential({ connector: id.value, credential });
    replaceOpen.value = false;
  } catch (e: unknown) {
    failureToast('Failed to replace token', e);
  }
}

async function handleRefresh() {
  try {
    await refresh({ connector: id.value });
  } catch (e: unknown) {
    failureToast('Failed to check connector', e);
  }
}

async function handleDelete() {
  try {
    await deleteConnector({ connector: id.value });
    successToast(`Connector "${props.name}" deleted`);
    router.push({ name: 'workspace', params: { drawer: 'connections' } });
  } catch (e: unknown) {
    failureToast('Failed to delete connector', e);
  }
}
</script>

<template>
  <ContextDrawer :title="connector?.name ?? name" @close="emit('close')">
    <template #leading>
      <IconButton label="Back to connections" class="-ml-1 shrink-0" as-child>
        <RouterLink :to="{ name: 'workspace', params: { drawer: 'connections' } }">
          <ArrowLeft />
        </RouterLink>
      </IconButton>
    </template>

    <template v-if="isAdmin && connector" #actions>
      <Button variant="outline" size="xs" :pending="refreshing" @click="handleRefresh">
        <RotateCw />
        {{ refreshing ? 'Checking…' : 'Check again' }}
      </Button>
    </template>

    <div v-if="loading && !result" class="space-y-6">
      <Skeleton class="h-4 w-2/3" />
      <Skeleton class="h-10 w-full" />
      <div class="space-y-2">
        <Skeleton v-for="i in 4" :key="i" class="h-10 w-full" />
      </div>
    </div>

    <QueryError
      v-else-if="error && !result"
      title="Couldn't load this connector"
      :error="error"
      @retry="refetch()"
    />

    <div v-else-if="connector" class="space-y-8">
      <p class="text-sm text-muted-foreground">
        MCP server at <span class="font-mono break-all">{{ connector.url }}</span>
      </p>

      <section class="space-y-2">
        <h3 class="text-sm font-medium text-muted-foreground">Token</h3>
        <div class="flex items-center gap-2 rounded-md bg-muted/40 py-2 pr-2 pl-3">
          <span v-if="connector.credentialHint" class="min-w-0 flex-1 truncate font-mono text-xs text-foreground">••••••••••••{{ connector.credentialHint }}</span>
          <span v-else class="min-w-0 flex-1 truncate text-xs text-muted-foreground">No token, the server needs no sign-in</span>
          <Status
            v-if="status && connector.status !== CredentialStatus.Valid"
            variant="text"
            :tone="status.tone"
            class="shrink-0 text-xs"
          >
            {{ status.label }}
          </Status>
          <Button v-if="isAdmin" variant="outline" size="xs" class="shrink-0" @click="replaceOpen = true">
            {{ connector.credentialHint ? 'Replace' : 'Add' }}
          </Button>
        </div>
        <p class="text-xs text-muted-foreground">
          <template v-if="connector.checkedAt">Checked <RelativeTime :value="connector.checkedAt" />. </template>
          Sent as a bearer token on every call. Agents never see it.
        </p>
      </section>

      <section class="space-y-3">
        <div class="flex items-baseline justify-between gap-4">
          <h3 class="text-sm font-medium text-muted-foreground">Tools</h3>
          <span class="text-xs text-muted-foreground tabular-nums">{{ connector.tools.length }}</span>
        </div>
        <p v-if="connector.tools.length === 0" class="text-sm text-muted-foreground">The server listed no tools at the last check.</p>
        <div v-else class="-mx-4 divide-y border-y">
          <div v-for="tool in connector.tools" :key="tool.name" class="space-y-0.5 px-4 py-2">
            <p class="font-mono text-sm text-foreground">{{ tool.name }}</p>
            <p v-if="tool.description" class="line-clamp-2 text-xs text-muted-foreground">{{ tool.description }}</p>
          </div>
        </div>
      </section>

      <section class="space-y-3">
        <div class="space-y-1 text-sm text-muted-foreground/80">
          <h3 class="text-sm font-medium text-muted-foreground">Agents</h3>
          <p>Agents granted this connector. Grant it from an agent's Access tab.</p>
        </div>
        <ConnectionAgents :agents="connector.agents" empty="No agent is granted this connector yet." />
      </section>

      <div v-if="isAdmin" class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3">
        <div>
          <p class="text-sm font-medium text-foreground">Delete connector</p>
          <p class="text-xs text-muted-foreground">Its token is deleted, and every agent loses its tools.</p>
        </div>
        <ConfirmDialog
          :title="`Delete connector &quot;${connector.name}&quot;?`"
          description="Its token is deleted, and it's removed from every agent it was granted to. This can't be undone."
          confirm-label="Delete connector"
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
      :title="`${connector?.credentialHint ? 'Replace' : 'Add'} the token of &quot;${connector?.name ?? name}&quot;`"
      description="The token is checked by listing the tools. The old one stays until the check passes."
      label="Token"
      :confirm-label="connector?.credentialHint ? 'Replace token' : 'Add token'"
      :pending="replacing"
      @submit="handleReplace"
    />
  </ContextDrawer>
</template>
