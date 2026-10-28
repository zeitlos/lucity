<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { ArrowLeft, RotateCw, Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import { CredentialStatus } from '@/gql/graphql';
import { modelApiLabel } from '@/lib/agents';
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

const ModelProviderDocument = graphql(`
  query ModelProvider($id: ModelProviderID!) {
    modelProvider(id: $id) {
      id
      name
      api
      url
      keyHint
      status
      checkedAt
      models
      agents {
        id
        name
        harness
      }
    }
  }
`);

const ReplaceModelProviderKeyDocument = graphql(`
  mutation ReplaceModelProviderKey($provider: ModelProviderID!, $apiKey: String!) {
    replaceModelProviderKey(provider: $provider, apiKey: $apiKey) {
      id
      keyHint
      status
      checkedAt
      models
    }
  }
`);

const RefreshModelProviderDocument = graphql(`
  mutation RefreshModelProvider($provider: ModelProviderID!) {
    refreshModelProvider(provider: $provider) {
      id
      status
      checkedAt
      models
    }
  }
`);

const DeleteModelProviderDocument = graphql(`
  mutation DeleteModelProvider($provider: ModelProviderID!) {
    deleteModelProvider(provider: $provider)
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

const { result, loading, error, refetch } = useQuery(ModelProviderDocument, () => ({ id: id.value }));
const provider = computed(() => result.value?.modelProvider ?? null);
const status = computed(() => (provider.value ? credentialStatus(provider.value.status, 'Key') : null));

const { mutate: replaceKey, loading: replacing } = useMutation(ReplaceModelProviderKeyDocument);
const { mutate: refresh, loading: refreshing } = useMutation(RefreshModelProviderDocument);
const { mutate: deleteProvider, loading: deleting } = useMutation(DeleteModelProviderDocument);

const replaceOpen = ref(false);

async function handleReplace(apiKey: string) {
  try {
    await replaceKey({ provider: id.value, apiKey });
    replaceOpen.value = false;
  } catch (e: unknown) {
    failureToast('Failed to replace key', e);
  }
}

async function handleRefresh() {
  try {
    await refresh({ provider: id.value });
  } catch (e: unknown) {
    failureToast('Failed to check model provider', e);
  }
}

async function handleDelete() {
  try {
    await deleteProvider({ provider: id.value });
    successToast(`Model provider "${props.name}" deleted`);
    router.push({ name: 'workspace', params: { drawer: 'connections' } });
  } catch (e: unknown) {
    failureToast('Failed to delete model provider', e);
  }
}
</script>

<template>
  <ContextDrawer :title="provider?.name ?? name" @close="emit('close')">
    <template #leading>
      <IconButton label="Back to connections" class="-ml-1 shrink-0" as-child>
        <RouterLink :to="{ name: 'workspace', params: { drawer: 'connections' } }">
          <ArrowLeft />
        </RouterLink>
      </IconButton>
    </template>

    <template v-if="isAdmin && provider" #actions>
      <Button variant="outline" size="xs" :pending="refreshing" @click="handleRefresh">
        <RotateCw />
        {{ refreshing ? 'Checking…' : 'Check again' }}
      </Button>
    </template>

    <div v-if="loading && !result" class="space-y-6">
      <Skeleton class="h-4 w-2/3" />
      <Skeleton class="h-10 w-full" />
      <div class="space-y-2">
        <Skeleton v-for="i in 3" :key="i" class="h-9 w-full" />
      </div>
    </div>

    <QueryError
      v-else-if="error && !result"
      title="Couldn't load this model provider"
      :error="error"
      @retry="refetch()"
    />

    <div v-else-if="provider" class="space-y-8">
      <p class="text-sm text-muted-foreground">
        {{ modelApiLabel(provider.api) }} at <span class="font-mono break-all">{{ provider.url }}</span>
      </p>

      <section class="space-y-2">
        <h3 class="text-sm font-medium text-muted-foreground">API key</h3>
        <div class="flex items-center gap-2 rounded-md bg-muted/40 py-2 pr-2 pl-3">
          <span class="min-w-0 flex-1 truncate font-mono text-xs text-foreground">••••••••••••{{ provider.keyHint ?? '' }}</span>
          <Status
            v-if="status && provider.status !== CredentialStatus.Valid"
            variant="text"
            :tone="status.tone"
            class="shrink-0 text-xs"
          >
            {{ status.label }}
          </Status>
          <Button v-if="isAdmin" variant="outline" size="xs" class="shrink-0" @click="replaceOpen = true">Replace</Button>
        </div>
        <p class="text-xs text-muted-foreground">
          <template v-if="provider.checkedAt">Checked <RelativeTime :value="provider.checkedAt" />. </template>
          The full key is never shown again.
        </p>
      </section>

      <section class="space-y-3">
        <div class="flex items-baseline justify-between gap-4">
          <h3 class="text-sm font-medium text-muted-foreground">Models</h3>
          <span class="text-xs text-muted-foreground tabular-nums">{{ provider.models.length }}</span>
        </div>
        <p v-if="provider.models.length === 0" class="text-sm text-muted-foreground">The provider listed no models at the last check.</p>
        <div v-else class="-mx-4 divide-y border-y">
          <div v-for="model in provider.models" :key="model" class="px-4 py-2 font-mono text-sm text-foreground">
            {{ model }}
          </div>
        </div>
      </section>

      <section class="space-y-3">
        <div class="space-y-1 text-sm text-muted-foreground/80">
          <h3 class="text-sm font-medium text-muted-foreground">Agents</h3>
          <p>Agents whose harness runs on these models.</p>
        </div>
        <ConnectionAgents :agents="provider.agents" empty="No agent uses these models yet." />
      </section>

      <div v-if="isAdmin" class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3">
        <div>
          <p class="text-sm font-medium text-foreground">Delete model provider</p>
          <p class="text-xs text-muted-foreground">The key is deleted. Agents lose these models right away.</p>
        </div>
        <ConfirmDialog
          :title="`Delete model provider &quot;${provider.name}&quot;?`"
          description="Its key is deleted, and running sessions on its models stop working. This can't be undone."
          confirm-label="Delete model provider"
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
      :title="`Replace the key of &quot;${provider?.name ?? name}&quot;`"
      description="The new key is checked by listing the models. The old one stays until the check passes."
      label="API key"
      confirm-label="Replace key"
      :pending="replacing"
      @submit="handleReplace"
    />
  </ContextDrawer>
</template>
