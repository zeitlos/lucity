<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core';
import { HardDrive, Plug } from '@lucide/vue';

defineProps<{
  data: {
    name: string;
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
  (e: 'mount'): void;
}>();
</script>

<template>
  <div class="volume-node-wrapper">
    <div
      :class="[
        'volume-node group relative cursor-pointer rounded-xl border px-6 py-5 shadow-sm transition-all duration-200',
        'hover:shadow-md',
        selected ? 'border-primary shadow-md' : 'border-border',
      ]"
      style="width: 280px;"
      @click="emit('select')"
    >
      <!-- Icon + name + hover mount button -->
      <div class="flex items-center gap-3">
        <HardDrive :size="26" class="shrink-0 text-muted-foreground" />
        <span class="min-w-0 flex-1 truncate font-semibold text-foreground">{{ data.name }}</span>
        <button
          class="pointer-events-none inline-flex shrink-0 items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground opacity-0 shadow-sm transition-opacity duration-150 hover:bg-primary/90 group-hover:pointer-events-auto group-hover:opacity-100"
          @click.stop="emit('mount')"
        >
          <Plug :size="14" />
          Mount
        </button>
      </div>
    </div>

    <!-- Vue Flow handles -->
    <Handle type="source" :position="Position.Bottom" class="!invisible" />
    <Handle type="target" :position="Position.Top" class="!invisible" />
  </div>
</template>

<style scoped>
.volume-node-wrapper {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.volume-node {
  z-index: 1;
  background: linear-gradient(
    to bottom,
    var(--card) 0%,
    color-mix(in oklch, var(--card) 94%, var(--muted)) 100%
  );
}
</style>
