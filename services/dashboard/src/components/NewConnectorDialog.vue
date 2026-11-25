<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { ExternalLink, Plug, Server } from '@lucide/vue';
import { graphql } from '@/gql';
import { connectorPresets, resourceName, type ConnectorPreset } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import BrandTile from '@/components/BrandTile.vue';
import ChoiceRows from '@/components/ChoiceRows.vue';
import DialogIcon from '@/components/DialogIcon.vue';
import SecretInput from '@/components/SecretInput.vue';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
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

const presetOptions = connectorPresets.map(option => ({ value: option.id, label: option.label, icon: option.id === 'mcp' ? Server : undefined }));

const preset = ref<ConnectorPreset>(connectorPresets[0]!);
const name = ref('');
const url = ref('');
const token = ref('');
const toolsetId = ref('');

const toolset = computed(() => preset.value.toolsets?.find(candidate => candidate.id === toolsetId.value) ?? null);

function applyPreset(next: ConnectorPreset) {
  preset.value = next;
  toolsetId.value = next.toolsets?.[0]?.id ?? '';
  url.value = next.url;
}

watch(open, (value) => {
  if (!value) return;
  applyPreset(connectorPresets[0]!);
  name.value = preset.value.name;
  token.value = '';
});

watch(toolset, (next) => {
  if (next) url.value = next.url;
});

function pick(id: string) {
  const next = connectorPresets.find(candidate => candidate.id === id);
  const previous = preset.value;
  if (!next || next === previous) return;
  applyPreset(next);
  if (!name.value.trim() || name.value === previous.name) name.value = next.name;
}

const asksUrl = computed(() => !!preset.value.urlPlaceholder);

function focusNextField() {
  const field = asksUrl.value && !url.value.trim() ? 'connector-url' : 'connector-token';
  document.getElementById(field)?.focus();
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
          <DialogDescription>A remote MCP server. Agents reach it through Lucity and never see the token.</DialogDescription>
        </DialogHeader>

        <fieldset :disabled="pending" class="min-w-0 space-y-4">
          <div class="space-y-2">
            <p id="connector-preset" class="text-sm font-medium text-foreground">Server</p>
            <ChoiceRows
              :model-value="preset.id"
              :options="presetOptions"
              aria-labelledby="connector-preset"
              @update:model-value="pick"
              @pointer="focusNextField"
            >
              <template #visual="{ option }">
                <BrandTile :name="option.label" :icon="option.icon" />
              </template>
            </ChoiceRows>
          </div>

          <div v-if="preset.toolsets" class="space-y-2">
            <Label for="connector-toolset">Tools</Label>
            <Select v-model="toolsetId">
              <SelectTrigger id="connector-toolset" class="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="option in preset.toolsets" :key="option.id" :value="option.id">
                  {{ option.label }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p v-if="toolset" class="text-xs text-muted-foreground">{{ toolset.hint }}</p>
          </div>

          <div v-if="asksUrl" class="space-y-2">
            <Label for="connector-url">URL</Label>
            <Input
              id="connector-url"
              v-model="url"
              class="font-mono text-sm"
              :placeholder="preset.urlPlaceholder"
              autocomplete="off"
              spellcheck="false"
            />
          </div>

          <div class="space-y-2">
            <div class="flex items-center justify-between gap-2">
              <Label for="connector-token">{{ preset.keyLabel }}</Label>
              <a
                v-if="preset.keyUrl"
                :href="preset.keyUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-1 rounded-sm text-xs text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
              >
                Get one
                <ExternalLink :size="12" />
              </a>
            </div>
            <SecretInput id="connector-token" v-model="token" />
            <p class="text-xs text-muted-foreground">
              {{ preset.keyHint ?? 'Sent as a bearer token.' }}<template v-if="preset.keyOptional"> Leave it empty for a server without sign-in.</template>
            </p>
          </div>

          <div class="space-y-2">
            <Label for="connector-name">Name</Label>
            <Input id="connector-name" v-model="name" :placeholder="preset.name || 'my-server'" autocomplete="off" spellcheck="false" />
            <p class="text-xs text-muted-foreground">Its tools are listed under this name.</p>
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
