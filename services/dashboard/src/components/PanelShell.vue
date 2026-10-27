<script setup lang="ts">
import { X } from '@lucide/vue';
import IconButton from '@/components/IconButton.vue';
import { useEscapeToClose } from '@/composables/useEscapeToClose';
import { confirmDiscard } from '@/composables/useUnsavedChanges';

const props = withDefaults(defineProps<{
  title: string;
  as?: 'section' | 'aside';
  guard?: boolean;
}>(), {
  as: 'section',
  guard: true,
});

const emit = defineEmits<{
  (e: 'close'): void;
}>();

async function close() {
  if (!props.guard || await confirmDiscard()) emit('close');
}

useEscapeToClose(close);
</script>

<template>
  <component
    :is="as"
    :aria-label="title"
    class="flex h-full flex-col rounded-lg border bg-card shadow-lg"
  >
    <header class="flex shrink-0 items-center gap-3 border-b px-4 py-3">
        <span
          v-if="$slots.icon"
          class="flex size-6 shrink-0 items-center justify-center text-foreground [&_img]:size-6 [&_svg]:size-6"
        >
          <slot name="icon" />
        </span>
        <h2 class="min-w-0 flex-1 truncate text-lg font-semibold text-foreground">{{ title }}</h2>
        <slot name="actions" />
        <IconButton
          label="Close"
          shortcut="Esc"
          @click="close"
        >
          <X />
        </IconButton>
    </header>

    <slot />

    <footer v-if="$slots.footer" class="shrink-0 border-t px-4 py-3">
      <slot name="footer" />
    </footer>
  </component>
</template>
