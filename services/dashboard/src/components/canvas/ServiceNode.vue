<script setup lang="ts">
import { computed } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import { Globe, Lock } from 'lucide-vue-next';
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
      'group cursor-pointer rounded-xl border bg-card px-4 py-3 shadow-sm transition-all duration-200',
      'hover:shadow-md',
      selected ? 'border-primary shadow-md' : 'border-border',
    ]"
    style="min-width: 220px;"
    @click="emit('select')"
  >
    <div class="flex items-center gap-3">
      <FrameworkIcon :framework="data.framework" :size="24" />
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <span class="truncate text-sm font-semibold text-foreground">{{ data.name }}</span>
          <span :class="['h-2 w-2 shrink-0 rounded-full', statusColor]" />
        </div>
        <div
          v-if="data.imageTag"
          class="mt-0.5 truncate text-xs text-muted-foreground"
        >
          {{ data.imageTag }}
        </div>
      </div>
      <component
        :is="data.public ? Globe : Lock"
        :size="14"
        class="shrink-0 text-muted-foreground"
      />
    </div>

    <!-- Vue Flow handles (invisible, for potential edges) -->
    <Handle type="source" :position="Position.Bottom" class="!invisible" />
    <Handle type="target" :position="Position.Top" class="!invisible" />
  </div>
</template>
