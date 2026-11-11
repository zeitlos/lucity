<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { computed, nextTick, ref } from 'vue';
import { Eye, EyeOff } from '@lucide/vue';
import IconButton from '@/components/IconButton.vue';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { cn } from '@/lib/utils';

defineOptions({ inheritAttrs: false });

const value = defineModel<string>({ required: true });
const revealed = defineModel<boolean>('revealed', { default: false });

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();

const multiline = computed(() => value.value.includes('\n'));
const textarea = ref<InstanceType<typeof Textarea> | null>(null);

async function insert(field: HTMLInputElement | HTMLTextAreaElement, text: string) {
  const start = field.selectionStart ?? value.value.length;
  const end = field.selectionEnd ?? value.value.length;
  value.value = value.value.slice(0, start) + text + value.value.slice(end);
  await nextTick();
  const element = textarea.value?.$el as HTMLTextAreaElement | undefined;
  element?.focus();
  element?.setSelectionRange(start + text.length, start + text.length);
}

function onPaste(event: ClipboardEvent) {
  const text = event.clipboardData?.getData('text') ?? '';
  if (!text.includes('\n')) return;
  event.preventDefault();
  insert(event.target as HTMLInputElement, text.replace(/\r\n?/g, '\n'));
}

function onKeydown(event: KeyboardEvent) {
  if (event.key !== 'Enter' || !(event.metaKey || event.ctrlKey)) return;
  event.preventDefault();
  insert(event.target as HTMLInputElement | HTMLTextAreaElement, '\n');
}
</script>

<template>
  <div class="relative">
    <Textarea
      v-if="multiline"
      ref="textarea"
      v-bind="$attrs"
      v-model="value"
      rows="1"
      autocomplete="off"
      autocapitalize="off"
      spellcheck="false"
      data-1p-ignore
      data-lpignore="true"
      data-bwignore
      :class="cn('max-h-64 min-h-9 resize-none pr-9 font-mono text-sm break-all [field-sizing:content]', !revealed && '[-webkit-text-security:disc]', props.class)"
      @keydown="onKeydown"
    />
    <Input
      v-else
      v-bind="$attrs"
      v-model="value"
      type="text"
      autocomplete="off"
      autocapitalize="off"
      spellcheck="false"
      data-1p-ignore
      data-lpignore="true"
      data-bwignore
      :class="cn('pr-9 font-mono text-sm', !revealed && value && '[-webkit-text-security:disc]', props.class)"
      @paste="onPaste"
      @keydown="onKeydown"
    />
    <IconButton
      :label="revealed ? 'Hide value' : 'Show value'"
      :aria-pressed="revealed"
      size="icon"
      tabindex="-1"
      class="absolute right-0 top-0 text-muted-foreground hover:bg-transparent"
      @click="revealed = !revealed"
    >
      <EyeOff v-if="revealed" />
      <Eye v-else />
    </IconButton>
  </div>
</template>
