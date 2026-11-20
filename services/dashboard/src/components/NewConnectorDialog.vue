<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { Plug } from '@lucide/vue';
import { graphql } from '@/gql';
import { connectorPresets, resourceName, type ConnectorPreset } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import BrandTile from '@/components/BrandTile.vue';
import DialogIcon from '@/components/DialogIcon.vue';
import SecretInput from '@/components/SecretInput.vue';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { failureToast } from '@/components/ui/sonner';

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

const preset = ref<ConnectorPreset>(connectorPresets[0]!);
const name = ref('');
const url = ref('');
const token = ref('');

watch(open, (value) => {
  if (value) pick(connectorPresets[0]!);
});

function pick(next: ConnectorPreset) {
  preset.value = next;
  name.value = next.name;
  url.value = next.url;
  token.value = '';
}

const { mutate: createConnector, loading: pending } = useMutation(CreateConnectorDocument);

const canCreate = computed(() => !pending.value && !!name.value.trim() && !!url.value.trim());

async function create() {
  if (!canCreate.value) return;
  try {
    const res = await createConnector({
      name: name.value.trim(),
      url: url.value.trim(),
      credential: token.value || null,
    });
    const id = res?.data?.createConnector.id;
    open.value = false;
    emit('created');
    if (id) router.push({ name: 'connection', params: { kind: 'connectors', name: resourceName(id) } });
  } catch (e: unknown) {
    failureToast('Failed to create connector', e);
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent :class="[mobileSheetClass, 'sm:max-w-md']">
      <form class="space-y-4" @submit.prevent="create">
        <DialogHeader>
          <DialogIcon :icon="Plug" />
          <DialogTitle>New connector</DialogTitle>
          <DialogDescription>A remote MCP server. Agents reach it through Lucity, which adds the token to every request.</DialogDescription>
        </DialogHeader>

        <fieldset :disabled="pending" class="min-w-0 space-y-4">
          <div class="space-y-2">
            <p class="text-sm font-medium text-foreground">Server</p>
            <div class="grid grid-cols-2 gap-1">
              <button
                v-for="option in connectorPresets"
                :key="option.id"
                type="button"
                :aria-pressed="option.id === preset.id"
                class="flex items-center gap-2 rounded-md px-2 py-2 text-left text-sm text-foreground transition-colors hover:bg-accent aria-pressed:bg-accent"
                @click="pick(option)"
              >
                <BrandTile :name="option.label" size="sm" />
                <span class="truncate">{{ option.label }}</span>
              </button>
            </div>
          </div>

          <div class="space-y-2">
            <Label for="connector-name">Name</Label>
            <Input id="connector-name" v-model="name" placeholder="linear" autocomplete="off" spellcheck="false" />
          </div>

          <div class="space-y-2">
            <Label for="connector-url">URL</Label>
            <Input
              id="connector-url"
              v-model="url"
              class="font-mono text-sm"
              placeholder="https://mcp.example.com/mcp"
              autocomplete="off"
              spellcheck="false"
            />
          </div>

          <div class="space-y-2">
            <Label for="connector-token">Token</Label>
            <SecretInput id="connector-token" v-model="token" />
            <p class="text-xs text-muted-foreground">Sent as a bearer token. Leave it empty for a server without sign-in.</p>
          </div>
        </fieldset>

        <DialogFooter>
          <Button type="button" variant="ghost" :disabled="pending" @click="open = false">Cancel</Button>
          <Button type="submit" :pending="pending" :disabled="!canCreate">
            {{ pending ? 'Creating…' : 'Create connector' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
