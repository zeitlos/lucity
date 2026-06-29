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
      <!-- Icon + name -->
      <div class="flex items-center gap-3">
        <HardDrive :size="26" class="shrink-0 text-muted-foreground" />
        <span class="truncate font-semibold text-foreground">{{ data.name }}</span>
      </div>

      <!-- Hover overlay: mount this volume -->
      <div
        class="pointer-events-none absolute inset-0 flex items-center justify-center rounded-xl bg-background/70 opacity-0 backdrop-blur-[1px] transition-opacity duration-150 group-hover:pointer-events-auto group-hover:opacity-100"
        @click="emit('select')"
      >
        <button
          class="inline-flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm transition-colors hover:bg-primary/90"
          @click.stop="emit('mount')"
        >
          <Plug :size="15" />
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
