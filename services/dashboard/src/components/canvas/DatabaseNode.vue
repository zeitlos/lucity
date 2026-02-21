<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core';
import { Database } from 'lucide-vue-next';
import { Badge } from '@/components/ui/badge';
import { Chip } from '@/components/ui/chip';

defineProps<{
  data: {
    name: string;
    version: string;
    instances: number;
    size: string;
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
}>();
</script>

<template>
  <div
    :class="[
      'database-node group cursor-pointer rounded-xl border px-6 py-5 shadow-sm transition-all duration-200',
      'hover:shadow-md',
      selected ? 'border-primary shadow-md' : 'border-border',
    ]"
    style="width: 280px;"
    @click="emit('select')"
  >
    <!-- Chip label -->
    <div class="mb-4 flex items-center justify-between">
      <Chip>postgres</Chip>
      <Badge variant="secondary" class="text-[0.65rem]">v{{ data.version }}</Badge>
    </div>

    <!-- Header: icon + name -->
    <div class="flex items-center gap-3">
      <Database :size="28" class="shrink-0 text-blue-500" />
      <span class="truncate font-semibold text-foreground">{{ data.name }}</span>
    </div>

    <!-- Meta row -->
    <div class="mt-4 flex items-center gap-3 border-t border-border/50 pt-4 text-xs text-muted-foreground">
      <span>{{ data.instances }} {{ data.instances === 1 ? 'instance' : 'instances' }}</span>
      <span class="ml-auto font-mono">{{ data.size }}</span>
    </div>

    <!-- Vue Flow handles -->
    <Handle type="source" :position="Position.Bottom" class="!invisible" />
    <Handle type="target" :position="Position.Top" class="!invisible" />
  </div>
</template>

<style scoped>
.database-node {
  background: linear-gradient(
    to bottom,
    var(--card) 0%,
    color-mix(in oklch, var(--card) 94%, var(--muted)) 100%
  );
}
</style>
