<script setup lang="ts">
import { computed } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import FrameworkIcon from '@/components/FrameworkIcon.vue';

const props = defineProps<{
  data: {
    name: string;
    framework?: string;
    port?: number;
    public?: boolean;
    ready?: boolean;
    imageTag?: string;
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
}>();

const statusColor = computed(() => {
  if (props.data.ready === undefined) return 'bg-muted-foreground/50';
  return props.data.ready ? 'bg-green-500' : 'bg-red-500';
});
</script>

<template>
  <div
    :class="[
      'group cursor-pointer rounded-xl border bg-card px-4 py-3.5 shadow-sm transition-all duration-200',
      'hover:shadow-md',
      selected ? 'border-primary shadow-md' : 'border-border',
    ]"
    style="width: 220px;"
    @click="emit('select')"
  >
    <!-- Header: icon + name -->
    <div class="flex items-center gap-2.5">
      <FrameworkIcon :framework="data.framework" :size="24" />
      <span class="truncate text-sm font-semibold text-foreground">{{ data.name }}</span>
    </div>

    <!-- URL / image tag -->
    <div
      v-if="data.imageTag"
      class="mt-1 truncate text-xs text-muted-foreground"
    >
      {{ data.imageTag }}
    </div>

    <!-- Status -->
    <div class="mt-3 flex items-center gap-1.5">
      <span :class="['h-2 w-2 shrink-0 rounded-full', statusColor]" />
      <span class="text-xs text-muted-foreground">
        {{ data.ready === undefined ? 'Unknown' : data.ready ? 'Online' : 'Not Ready' }}
      </span>
    </div>

    <!-- Vue Flow handles (invisible, for potential edges) -->
    <Handle type="source" :position="Position.Bottom" class="!invisible" />
    <Handle type="target" :position="Position.Top" class="!invisible" />
  </div>
</template>
