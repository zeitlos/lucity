<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { Cable, Plug, Sparkles } from '@lucide/vue';
import { graphql } from '@/gql';
import { ModelApi } from '@/gql/graphql';
import {
  connectorPresets,
  modelProviderPresets,
  resourceName,
  type ConnectorPreset,
  type ModelProviderPreset,
} from '@/lib/agents';
import BrandTile from '@/components/BrandTile.vue';
import DialogIcon from '@/components/DialogIcon.vue';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { failureToast } from '@/components/ui/sonner';

const CreateModelProviderDocument = graphql(`
  mutation CreateModelProvider($name: String!, $api: ModelApi!, $url: String, $apiKey: String!) {
    createModelProvider(name: $name, api: $api, url: $url, apiKey: $apiKey) {
      id
      name
      api
      url
      keyHint
      status
      checkedAt
      models
    }
  }
`);

const CreateConnectorDocument = graphql(`
  mutation CreateConnector($name: String!, $url: String!, $credential: String) {
    createConnector(name: $name, url: $url, credential: $credential) {
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
    }
  }
`);

const open = defineModel<boolean>('open', { default: false });

const emit = defineEmits<{
  (e: 'created'): void;
}>();

const router = useRouter();

const provider = ref<ModelProviderPreset | null>(null);
const connector = ref<ConnectorPreset | null>(null);

const name = ref('');
const url = ref('');
const secret = ref('');

watch(open, (value) => {
  if (value) back();
});

function back() {
  provider.value = null;
  connector.value = null;
  name.value = '';
  url.value = '';
  secret.value = '';
}

function pickProvider(preset: ModelProviderPreset) {
  provider.value = preset;
  name.value = preset.name;
  url.value = preset.url;
}

function pickConnector(preset: ConnectorPreset) {
  connector.value = preset;
  name.value = preset.name;
  url.value = preset.url;
}

const needsUrl = computed(() => provider.value?.api === ModelApi.OpenaiCompatible);

const { mutate: createModelProvider, loading: creatingProvider } = useMutation(CreateModelProviderDocument);
const { mutate: createConnector, loading: creatingConnector } = useMutation(CreateConnectorDocument);

const pending = computed(() => creatingProvider.value || creatingConnector.value);

const canCreate = computed(() => {
  if (pending.value || !name.value.trim()) return false;
  if (provider.value) return !!secret.value && (!needsUrl.value || !!url.value.trim());
  if (connector.value) return !!url.value.trim();
  return false;
});

async function create() {
  if (!canCreate.value) return;
  try {
    if (provider.value) {
      const res = await createModelProvider({
        name: name.value.trim(),
        api: provider.value.api,
        url: needsUrl.value ? url.value.trim() : null,
        apiKey: secret.value,
      });
      const id = res?.data?.createModelProvider.id;
      open.value = false;
      emit('created');
      if (id) router.push({ name: 'connection', params: { kind: 'model-providers', name: resourceName(id) } });
    } else if (connector.value) {
      const res = await createConnector({
        name: name.value.trim(),
        url: url.value.trim(),
        credential: secret.value || null,
      });
      const id = res?.data?.createConnector.id;
      open.value = false;
      emit('created');
      if (id) router.push({ name: 'connection', params: { kind: 'connectors', name: resourceName(id) } });
    }
  } catch (e: unknown) {
    failureToast(provider.value ? 'Failed to create model provider' : 'Failed to create connector', e);
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="sm:max-w-md">
      <template v-if="!provider && !connector">
        <DialogHeader>
          <DialogIcon :icon="Cable" />
          <DialogTitle>New connection</DialogTitle>
          <DialogDescription>
            Connect it once for the workspace. Agents use it without ever seeing its key.
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-5">
          <div class="space-y-2">
            <p class="text-sm font-medium text-muted-foreground">Model providers</p>
            <div class="grid grid-cols-2 gap-1">
              <button
                v-for="preset in modelProviderPresets"
                :key="preset.id"
                type="button"
                class="flex items-center gap-2 rounded-md px-2 py-2 text-left text-sm text-foreground transition-colors hover:bg-accent"
                @click="pickProvider(preset)"
              >
                <BrandTile :name="preset.label" size="sm" />
                <span class="truncate">{{ preset.label }}</span>
              </button>
            </div>
          </div>

          <div class="space-y-2">
            <p class="text-sm font-medium text-muted-foreground">Connectors</p>
            <div class="grid grid-cols-2 gap-1">
              <button
                v-for="preset in connectorPresets"
                :key="preset.id"
                type="button"
                class="flex items-center gap-2 rounded-md px-2 py-2 text-left text-sm text-foreground transition-colors hover:bg-accent"
                @click="pickConnector(preset)"
              >
                <BrandTile :name="preset.label" size="sm" />
                <span class="truncate">{{ preset.label }}</span>
              </button>
            </div>
          </div>
        </div>
      </template>

      <form v-else class="space-y-4" @submit.prevent="create">
        <DialogHeader>
          <DialogIcon :icon="provider ? Sparkles : Plug" />
          <DialogTitle>{{ provider ? 'New model provider' : 'New connector' }}</DialogTitle>
          <DialogDescription v-if="provider">
            {{ provider.label }}. Its models show up wherever someone picks a model for a session.
          </DialogDescription>
          <DialogDescription v-else>
            A remote MCP server. Agents reach it through Lucity, which adds the token to every request.
          </DialogDescription>
        </DialogHeader>

        <fieldset :disabled="pending" class="min-w-0 space-y-4">
          <div class="space-y-2">
            <Label for="connection-name">Name</Label>
            <Input
              id="connection-name"
              v-model="name"
              :placeholder="provider ? 'anthropic' : 'linear'"
              autocomplete="off"
              spellcheck="false"
            />
          </div>

          <div v-if="connector || needsUrl" class="space-y-2">
            <Label for="connection-url">URL</Label>
            <Input
              id="connection-url"
              v-model="url"
              class="font-mono text-sm"
              :placeholder="provider?.urlPlaceholder ?? 'https://mcp.example.com/mcp'"
              autocomplete="off"
              spellcheck="false"
            />
            <p v-if="provider" class="text-xs text-muted-foreground">The base URL that serves /models and /chat/completions.</p>
          </div>

          <div class="space-y-2">
            <Label for="connection-secret">{{ provider ? provider.keyLabel : 'Token' }}</Label>
            <Input
              id="connection-secret"
              v-model="secret"
              type="password"
              class="font-mono text-sm"
              autocomplete="off"
              data-1p-ignore
            />
            <p class="text-xs text-muted-foreground">
              <template v-if="provider">Checked by listing the models. Agents never see it.</template>
              <template v-else>Sent as a bearer token. Leave it empty for a server without sign-in.</template>
            </p>
          </div>
        </fieldset>

        <DialogFooter>
          <Button type="button" variant="ghost" :disabled="pending" @click="back">Back</Button>
          <Button type="submit" :pending="pending" :disabled="!canCreate">
            <template v-if="pending">Creating…</template>
            <template v-else>{{ provider ? 'Create model provider' : 'Create connector' }}</template>
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
