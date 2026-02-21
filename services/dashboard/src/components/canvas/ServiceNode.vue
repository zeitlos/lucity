<script setup lang="ts">
import { computed } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import { Github, Globe } from 'lucide-vue-next';
import FrameworkIcon from '@/components/FrameworkIcon.vue';
import { Badge } from '@/components/ui/badge';
import { Chip } from '@/components/ui/chip';

const props = defineProps<{
  data: {
    name: string;
    framework?: string;
    port?: number;
    sourceUrl?: string;
    host?: string;
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

const shortRepoName = computed(() => {
  if (!props.data.sourceUrl) return null;
  return props.data.sourceUrl.replace('https://github.com/', '');
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

    <!-- Meta row: port + repo + domain + replicas -->
    <div class="mt-4 flex items-center gap-3 border-t border-border/50 pt-4 text-xs text-muted-foreground">
      <span v-if="data.port" class="font-mono">:{{ data.port }}</span>
      <span v-if="shortRepoName" class="flex items-center gap-1 truncate">
        <Github :size="12" class="shrink-0" />
        <span class="truncate">{{ shortRepoName }}</span>
      </span>
      <span v-if="data.host" class="flex items-center gap-1 truncate">
        <Globe :size="12" class="shrink-0" />
        <span class="truncate">{{ data.host }}</span>
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
