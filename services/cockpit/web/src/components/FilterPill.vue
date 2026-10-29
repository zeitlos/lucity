<script setup lang="ts">
import { ChevronDown, X } from '@lucide/vue';

defineProps<{
  id: string;
  label: string;
  summary: string;
}>();

const emit = defineEmits<{
  clear: [];
}>();
</script>

<template>
  <span
    class="flex h-8 items-center rounded-sm border"
    :class="summary ? 'border-foreground/25 bg-selected' : 'bg-background'"
    :style="`anchor-name: --${id}`"
  >
    <button type="button" :popovertarget="id" class="flex h-full items-center gap-1.5 pl-2.5" :class="summary ? 'pr-1' : 'pr-2'">
      <span :class="summary ? 'text-muted-foreground' : ''">{{ label }}</span>
      <span v-if="summary" class="max-w-48 truncate">{{ summary }}</span>
      <ChevronDown v-else class="size-3.5 text-muted-foreground" aria-hidden="true" />
    </button>
    <button
      v-if="summary"
      type="button"
      class="flex size-7 items-center justify-center text-muted-foreground hover:text-foreground"
      :aria-label="`Clear ${label} filter`"
      @click="emit('clear')"
    >
      <X class="size-3.5" aria-hidden="true" />
    </button>
    <div
      :id="id"
      popover
      class="m-0 mt-1 rounded-lg border bg-card p-1 text-foreground shadow-lg [position-area:bottom_span-right] [position-try-fallbacks:flip-inline]"
      :style="`position-anchor: --${id}`"
    >
      <slot />
    </div>
  </span>
</template>
