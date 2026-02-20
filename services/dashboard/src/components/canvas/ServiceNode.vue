<script setup lang="ts">
import { computed } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import { Globe, Lock } from 'lucide-vue-next';
import FrameworkIcon from '@/components/FrameworkIcon.vue';
import { Badge } from '@/components/ui/badge';
import { Chip } from '@/components/ui/chip';

const props = defineProps<{
  data: {
    name: string;
    framework?: string;
    port?: number;
    public?: boolean;
    ready?: boolean;
    imageTag?: string;
    replicas?: number;
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
}>();

const badgeVariant = computed(() => {
  if (props.data.ready === undefined) return 'secondary' as const;
  return props.data.ready ? 'default' as const : 'destructive' as const;
});

const statusLabel = computed(() => {
  if (props.data.ready === undefined) return 'Unknown';
  return props.data.ready ? 'Online' : 'Not Ready';
});
</script>

<template>
  <div
    :class="[
      'service-node group cursor-pointer rounded-xl border px-6 py-5 shadow-sm transition-all duration-200',
      'hover:shadow-md',
      selected ? 'border-primary shadow-md' : 'border-border',
    ]"
    style="width: 280px;"
    @click="emit('select')"
  >
    <!-- Chip label -->
    <div class="mb-4 flex items-center justify-between">
      <Chip v-if="data.framework">{{ data.framework }}</Chip>
      <Chip v-else>service</Chip>
      <Badge :variant="badgeVariant" class="text-[0.65rem]">{{ statusLabel }}</Badge>
    </div>

    <!-- Header: icon + name -->
    <div class="flex items-center gap-3">
      <FrameworkIcon :framework="data.framework" :size="28" />
      <span class="truncate font-semibold text-foreground">{{ data.name }}</span>
    </div>

    <!-- Meta row: port + visibility + replicas -->
    <div class="mt-4 flex items-center gap-3 border-t border-border/50 pt-4 text-xs text-muted-foreground">
      <span v-if="data.port" class="font-mono">:{{ data.port }}</span>
      <span class="flex items-center gap-1">
        <Globe v-if="data.public" :size="12" />
        <Lock v-else :size="12" />
        {{ data.public ? 'Public' : 'Private' }}
      </span>
      <span v-if="data.replicas !== undefined" class="ml-auto">
        {{ data.replicas }} {{ data.replicas === 1 ? 'replica' : 'replicas' }}
      </span>
    </div>

    <!-- Vue Flow handles (invisible, for potential edges) -->
    <Handle type="source" :position="Position.Bottom" class="!invisible" />
    <Handle type="target" :position="Position.Top" class="!invisible" />
  </div>
</template>

<style scoped>
.service-node {
  background: linear-gradient(
    to bottom,
    var(--card) 0%,
    color-mix(in oklch, var(--card) 94%, var(--muted)) 100%
  );
}
</style>
