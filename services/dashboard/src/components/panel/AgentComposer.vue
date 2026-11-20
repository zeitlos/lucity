<script setup lang="ts">
import { computed, ref } from 'vue';
import { ArrowUp, Square } from '@lucide/vue';
import type { Effort } from '@/gql/graphql';
import { efforts, modelKey, type AgentModelOption } from '@/lib/agents';
import BrandTile from '@/components/BrandTile.vue';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

const props = defineProps<{
  agentName: string;
  models: AgentModelOption[];
  model: AgentModelOption | null;
  effort: Effort;
  working: boolean;
  sending: boolean;
  stopping: boolean;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  (e: 'send', message: string): void;
  (e: 'stop'): void;
  (e: 'change', model: AgentModelOption, effort: Effort): void;
}>();

const draft = ref('');

const providers = computed(() => {
  const groups = new Map<string, { id: string; name: string; models: AgentModelOption[] }>();
  for (const model of props.models) {
    const group = groups.get(model.provider.id) ?? { id: model.provider.id, name: model.provider.name, models: [] };
    group.models.push(model);
    groups.set(model.provider.id, group);
  }
  return [...groups.values()];
});

const selectedKey = computed(() => (props.model ? modelKey(props.model) : undefined));

function changeModel(value: unknown) {
  const model = props.models.find(candidate => modelKey(candidate) === value);
  if (model) emit('change', model, props.effort);
}

function changeEffort(value: unknown) {
  const effort = efforts.find(option => option.value === value)?.value;
  if (effort && props.model) emit('change', props.model, effort);
}

const canSend = computed(() => draft.value.trim().length > 0 && !!props.model && !props.sending && !props.disabled);

function send() {
  if (!canSend.value) return;
  emit('send', draft.value.trim());
}

function clear() {
  draft.value = '';
}

function onKeydown(event: KeyboardEvent) {
  if (event.key !== 'Enter' || event.shiftKey || event.isComposing) return;
  event.preventDefault();
  send();
}

defineExpose({ clear });
</script>

<template>
  <form class="rounded-lg border bg-card shadow-xs focus-within:border-ring" @submit.prevent="send">
    <textarea
      v-model="draft"
      rows="3"
      :placeholder="`Message ${agentName}…`"
      :aria-label="`Message ${agentName}`"
      :disabled="disabled"
      class="block max-h-60 min-h-16 w-full resize-none bg-transparent px-3 pt-3 pb-1 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed"
      @keydown="onKeydown"
    />
    <div class="flex items-center gap-2 p-2">
      <Select :model-value="selectedKey" :disabled="disabled || models.length === 0" @update:model-value="changeModel">
        <SelectTrigger class="h-7 w-auto max-w-56 gap-2 border-0 bg-muted/60 px-2 text-xs shadow-none" aria-label="Model">
          <span v-if="model" class="flex min-w-0 items-center gap-1.5">
            <BrandTile :name="model.provider.name" size="sm" />
            <span class="truncate font-mono">{{ model.name }}</span>
          </span>
          <span v-else class="text-muted-foreground">No model</span>
        </SelectTrigger>
        <SelectContent>
          <SelectGroup v-for="provider in providers" :key="provider.id">
            <SelectLabel class="flex items-center gap-2">
              <BrandTile :name="provider.name" size="sm" />
              {{ provider.name }}
            </SelectLabel>
            <SelectItem
              v-for="option in provider.models"
              :key="modelKey(option)"
              :value="modelKey(option)"
              class="font-mono text-xs"
            >
              {{ option.name }}
            </SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>

      <Select :model-value="effort" :disabled="disabled || !model" @update:model-value="changeEffort">
        <SelectTrigger class="h-7 w-auto gap-2 border-0 bg-muted/60 px-2 text-xs shadow-none" aria-label="Effort">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="option in efforts" :key="option.value" :value="option.value" class="text-xs">
            {{ option.label }} effort
          </SelectItem>
        </SelectContent>
      </Select>

      <div class="flex-1" />

      <Button
        v-if="working"
        type="button"
        variant="outline"
        size="xs"
        :pending="stopping"
        @click="emit('stop')"
      >
        <Square />
        {{ stopping ? 'Stopping…' : 'Stop' }}
      </Button>
      <Button type="submit" size="xs" :pending="sending" :disabled="!canSend">
        <ArrowUp />
        {{ sending ? 'Sending…' : 'Send' }}
      </Button>
    </div>
  </form>
</template>
