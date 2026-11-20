<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { Sparkles } from '@lucide/vue';
import { graphql } from '@/gql';
import { ModelApi } from '@/gql/graphql';
import { modelProviderPresets, resourceName, type ModelProviderPreset } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import BrandTile from '@/components/BrandTile.vue';
import DialogIcon from '@/components/DialogIcon.vue';
import SecretInput from '@/components/SecretInput.vue';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
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

const open = defineModel<boolean>('open', { default: false });

const emit = defineEmits<{
  (e: 'created'): void;
}>();

const router = useRouter();

const preset = ref<ModelProviderPreset>(modelProviderPresets[0]!);
const name = ref('');
const url = ref('');
const key = ref('');

watch(open, (value) => {
  if (value) pick(modelProviderPresets[0]!);
});

function pick(next: ModelProviderPreset) {
  preset.value = next;
  name.value = next.name;
  url.value = next.url;
  key.value = '';
}

const needsUrl = computed(() => preset.value.api === ModelApi.OpenaiCompatible);

const { mutate: createModelProvider, loading: pending } = useMutation(CreateModelProviderDocument);

const canCreate = computed(() => !pending.value && !!name.value.trim() && !!key.value && (!needsUrl.value || !!url.value.trim()));

async function create() {
  if (!canCreate.value) return;
  try {
    const res = await createModelProvider({
      name: name.value.trim(),
      api: preset.value.api,
      url: needsUrl.value ? url.value.trim() : null,
      apiKey: key.value,
    });
    const id = res?.data?.createModelProvider.id;
    open.value = false;
    emit('created');
    if (id) router.push({ name: 'connection', params: { kind: 'model-providers', name: resourceName(id) } });
  } catch (e: unknown) {
    failureToast('Failed to create model provider', e);
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent :class="[mobileSheetClass, 'sm:max-w-md']">
      <form class="space-y-4" @submit.prevent="create">
        <DialogHeader>
          <DialogIcon :icon="Sparkles" />
          <DialogTitle>New model provider</DialogTitle>
          <DialogDescription>Its models show up wherever someone picks a model for a session.</DialogDescription>
        </DialogHeader>

        <fieldset :disabled="pending" class="min-w-0 space-y-4">
          <div class="space-y-2">
            <p class="text-sm font-medium text-foreground">Provider</p>
            <div class="grid grid-cols-2 gap-1">
              <button
                v-for="option in modelProviderPresets"
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
            <Label for="model-provider-name">Name</Label>
            <Input id="model-provider-name" v-model="name" placeholder="anthropic" autocomplete="off" spellcheck="false" />
          </div>

          <div v-if="needsUrl" class="space-y-2">
            <Label for="model-provider-url">URL</Label>
            <Input
              id="model-provider-url"
              v-model="url"
              class="font-mono text-sm"
              :placeholder="preset.urlPlaceholder ?? 'https://models.example.com/v1'"
              autocomplete="off"
              spellcheck="false"
            />
            <p class="text-xs text-muted-foreground">The base URL that serves /models and /chat/completions.</p>
          </div>

          <div class="space-y-2">
            <Label for="model-provider-key">{{ preset.keyLabel }}</Label>
            <SecretInput id="model-provider-key" v-model="key" />
            <p class="text-xs text-muted-foreground">Checked by listing the models. Agents never see it.</p>
          </div>
        </fieldset>

        <DialogFooter>
          <Button type="button" variant="ghost" :disabled="pending" @click="open = false">Cancel</Button>
          <Button type="submit" :pending="pending" :disabled="!canCreate">
            {{ pending ? 'Creating…' : 'Create model provider' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
