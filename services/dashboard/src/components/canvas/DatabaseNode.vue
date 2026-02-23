<script setup lang="ts">
import { computed } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import { HardDrive } from 'lucide-vue-next';
import { Badge } from '@/components/ui/badge';

const props = defineProps<{
  data: {
    name: string;
    version: string;
    instances: number;
    size: string;
    ready?: boolean;
    volume?: {
      name: string;
      size: string;
      requestedSize: string;
      usedBytes: number;
      capacityBytes: number;
    } | null;
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
  (e: 'select-volume', volumeName: string): void;
}>();

const badgeVariant = computed(() => {
  if (props.data.ready === undefined) return 'secondary' as const;
  return props.data.ready ? 'default' as const : 'destructive' as const;
});

const statusLabel = computed(() => {
  if (props.data.ready === undefined) return 'Unknown';
  return props.data.ready ? 'Online' : 'Not Ready';
});

const instances = computed(() => props.data.instances ?? 0);

const usagePercent = computed(() => {
  if (!props.data.volume || props.data.volume.capacityBytes <= 0) return 0;
  return Math.min(100, Math.round((props.data.volume.usedBytes / props.data.volume.capacityBytes) * 100));
});

const usageLabel = computed(() => {
  if (!props.data.volume || props.data.volume.capacityBytes <= 0) {
    return props.data.volume?.size || props.data.size;
  }
  return `${usagePercent.value}% of ${props.data.volume.size}`;
});
</script>

<template>
  <div class="database-node-wrapper">
    <!-- Main card -->
    <div
      :class="[
        'database-node group cursor-pointer rounded-xl border px-6 py-5 shadow-sm transition-all duration-200',
        'hover:shadow-md',
        selected ? 'border-primary shadow-md' : 'border-border',
        instances >= 2 && 'has-stack',
        instances >= 3 && 'has-stack-deep',
      ]"
      style="width: 280px;"
      @click="emit('select')"
    >
      <!-- Header: icon + name -->
      <div class="flex items-center gap-3">
        <img
          src="https://devicons.railway.com/i/postgresql.svg"
          :width="28"
          :height="28"
          class="shrink-0"
          alt=""
        />
        <span class="truncate font-semibold text-foreground">{{ data.name }}</span>
      </div>

      <!-- Version -->
      <div class="mt-3">
        <div class="flex items-center gap-1.5 text-xs text-muted-foreground">
          <span class="font-mono">PostgreSQL {{ data.version }}</span>
        </div>
      </div>

      <!-- Status row -->
      <div class="mt-4 flex items-center justify-between border-t border-border/50 pt-4">
        <Badge :variant="badgeVariant" class="text-[0.65rem]">{{ statusLabel }}</Badge>
        <span class="text-[0.65rem] font-mono text-muted-foreground">{{ data.size }}</span>
      </div>
    </div>

    <!-- Volume sub-element -->
    <div
      v-if="data.volume"
      class="volume-bar relative mt-2 flex cursor-pointer items-center gap-2 overflow-hidden rounded-lg border border-border/70 px-3 py-2 text-xs text-muted-foreground transition-colors hover:border-border hover:bg-card/80"
      style="width: 240px; margin-left: 20px;"
      @click.stop="emit('select-volume', data.volume.name)"
    >
      <div
        v-if="usagePercent > 0"
        class="usage-fill absolute inset-y-0 left-0"
        :style="{ width: usagePercent + '%' }"
      />
      <HardDrive :size="12" class="relative z-10 shrink-0" />
      <span class="relative z-10">Volume</span>
      <span class="relative z-10 ml-auto font-mono">{{ usageLabel }}</span>
    </div>

    <!-- Vue Flow handles -->
    <Handle type="source" :position="Position.Bottom" class="!invisible" />
    <Handle type="target" :position="Position.Top" class="!invisible" />
  </div>
</template>

<style scoped>
.database-node-wrapper {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.database-node {
  position: relative;
  z-index: 1;
  background: linear-gradient(
    to bottom,
    var(--card) 0%,
    color-mix(in oklch, var(--card) 94%, var(--muted)) 100%
  );
}

/* First stacked card (2+ instances) */
.has-stack::before {
  content: '';
  position: absolute;
  inset: 0;
  z-index: -1;
  border-radius: inherit;
  border: 1px solid var(--border);
  background: var(--card);
  transform: translateY(6px) scale(0.97);
  opacity: 0.7;
}

/* Second stacked card (3+ instances) */
.has-stack-deep::after {
  content: '';
  position: absolute;
  inset: 0;
  z-index: -2;
  border-radius: inherit;
  border: 1px solid var(--border);
  background: var(--card);
  transform: translateY(12px) scale(0.94);
  opacity: 0.4;
}

.volume-bar {
  background: color-mix(in oklch, var(--card) 60%, transparent);
  backdrop-filter: blur(4px);
}

.usage-fill {
  background: color-mix(in oklch, var(--primary) 15%, transparent);
}
</style>
