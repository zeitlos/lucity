<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { RadioGroupIndicator, RadioGroupItem } from 'reka-ui';
import { Check, ExternalLink, Sparkles } from '@lucide/vue';
import { graphql } from '@/gql';
import { ModelApi } from '@/gql/graphql';
import { modelProviderPresets, resourceName, type ModelProviderPreset } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import DialogIcon from '@/components/DialogIcon.vue';
import ModelProviderIcon from '@/components/ModelProviderIcon.vue';
import SecretInput from '@/components/SecretInput.vue';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { RadioGroup } from '@/components/ui/radio-group';
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
  if (!value) return;
  preset.value = modelProviderPresets[0]!;
  name.value = preset.value.name;
  url.value = '';
  key.value = '';
});

function pick(id: string) {
  const next = modelProviderPresets.find(candidate => candidate.id === id);
  const previous = preset.value;
  if (!next || next === previous) return;
  preset.value = next;
  if (!name.value.trim() || name.value === previous.name) name.value = next.name;
  if (next.url || previous.url) url.value = next.url;
}

const needsUrl = computed(() => preset.value.api === ModelApi.OpenaiCompatible);
const asksUrl = computed(() => needsUrl.value && !preset.value.url);
const keyHint = computed(() => (preset.value.keyHint ? `${preset.value.keyHint} Agents never see it.` : 'Agents never see it.'));

async function pickByPointer(event: MouseEvent, id: string) {
  if (!event.detail) return;
  pick(id);
  await nextTick();
  const field = asksUrl.value && !url.value.trim() ? 'model-provider-url' : 'model-provider-key';
  document.getElementById(field)?.focus();
}

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
          <DialogDescription>A provider's models show up wherever someone picks a model for a session.</DialogDescription>
        </DialogHeader>

        <fieldset :disabled="pending" class="min-w-0 space-y-4">
          <div class="space-y-2">
            <p id="model-provider-preset" class="text-sm font-medium text-foreground">Provider</p>
            <RadioGroup
              :model-value="preset.id"
              aria-labelledby="model-provider-preset"
              class="grid-cols-2 gap-1"
              @update:model-value="pick($event as string)"
            >
              <RadioGroupItem
                v-for="option in modelProviderPresets"
                :key="option.id"
                :value="option.id"
                class="flex items-center gap-2.5 rounded-md border border-transparent p-1 text-left text-sm text-foreground transition-colors hover:bg-accent/60 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none data-[state=checked]:border-border data-[state=checked]:bg-accent data-[state=checked]:font-medium"
                @click="pickByPointer($event, option.id)"
              >
                <ModelProviderIcon :brand="option.brand" />
                <span class="min-w-0 flex-1 truncate">{{ option.label }}</span>
                <RadioGroupIndicator class="mr-1 shrink-0">
                  <Check :size="16" />
                </RadioGroupIndicator>
              </RadioGroupItem>
            </RadioGroup>
          </div>

          <div v-if="asksUrl" class="space-y-2">
            <Label for="model-provider-url">URL</Label>
            <Input
              id="model-provider-url"
              v-model="url"
              class="font-mono text-sm"
              :placeholder="preset.urlPlaceholder"
              autocomplete="off"
              spellcheck="false"
            />
            <p class="text-xs text-muted-foreground">{{ preset.urlHint ?? 'The base URL that serves /models and /chat/completions.' }}</p>
          </div>

          <div class="space-y-2">
            <div class="flex items-center justify-between gap-2">
              <Label for="model-provider-key">{{ preset.keyLabel }}</Label>
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
            <SecretInput id="model-provider-key" v-model="key" :placeholder="preset.keyPlaceholder" />
            <p class="text-xs text-muted-foreground">{{ keyHint }}</p>
          </div>

          <div class="space-y-2">
            <Label for="model-provider-name">Name</Label>
            <Input id="model-provider-name" v-model="name" :placeholder="preset.name || 'my-provider'" autocomplete="off" spellcheck="false" />
            <p class="text-xs text-muted-foreground">Its models are listed under this name.</p>
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
