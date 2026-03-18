<script setup lang="ts">
import { ref } from 'vue';
import { Clipboard, Check } from 'lucide-vue-next';

const props = defineProps<{
  title: string;
  description?: string;
  copyText: string;
  action?: { label: string; onClick: (e: MouseEvent) => void };
}>();

const copied = ref(false);

async function handleCopy() {
  await navigator.clipboard.writeText(props.copyText);
  copied.value = true;
  setTimeout(() => { copied.value = false; }, 2000);
}
</script>

<template>
  <div class="flex items-start gap-1">
    <div class="flex-1 min-w-0">
      <div data-title class="text-sm font-medium">
        {{ title }}
      </div>
      <div
        v-if="description"
        data-description
        class="mt-1 text-xs text-muted-foreground"
      >
        {{ description }}
      </div>
      <button
        v-if="action"
        data-button
        data-action
        class="mt-2"
        @click="action.onClick"
      >
        {{ action.label }}
      </button>
    </div>
    <button
      class="shrink-0 rounded p-1 text-muted-foreground/60 transition-colors hover:text-muted-foreground"
      @click="handleCopy"
    >
      <Check v-if="copied" :size="14" class="text-[var(--status-ok)]" />
      <Clipboard v-else :size="14" />
    </button>
  </div>
</template>
