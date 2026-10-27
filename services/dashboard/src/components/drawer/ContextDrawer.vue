<script setup lang="ts">
import { X } from '@lucide/vue';
import { Button } from '@/components/ui/button';

defineProps<{
  title: string;
  description?: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();
</script>

<template>
  <aside
    :aria-label="title"
    class="flex h-full flex-col rounded-lg border bg-card shadow-lg"
  >
    <div class="flex shrink-0 items-center justify-between pb-1 pl-5 pr-4 pt-4">
      <h2 class="text-base font-semibold text-foreground">{{ title }}</h2>
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        aria-label="Close"
        @click="emit('close')"
      >
        <X :size="16" />
      </Button>
    </div>
    <p v-if="description" class="shrink-0 px-5 text-[13px] leading-5 text-muted-foreground">
      {{ description }}
    </p>

    <div class="min-h-0 flex-1 overflow-y-auto px-5 py-5">
      <slot />
    </div>

    <div v-if="$slots.footer" class="shrink-0 border-t px-5 py-3.5">
      <slot name="footer" />
    </div>
  </aside>
</template>
