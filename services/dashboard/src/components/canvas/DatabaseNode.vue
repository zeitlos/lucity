<script setup lang="ts">
import { computed } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import { HardDrive } from 'lucide-vue-next';
import { Badge } from '@/components/ui/badge';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';

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
    } | null;
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

const instances = computed(() => props.data.instances ?? 0);
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
    <Popover v-if="data.volume">
      <PopoverTrigger as-child>
        <div
          class="volume-bar mt-2 flex cursor-pointer items-center gap-2 rounded-lg border border-border/70 px-3 py-2 text-xs text-muted-foreground transition-colors hover:border-border hover:bg-card/80"
          style="width: 240px; margin-left: 20px;"
          @click.stop
        >
          <HardDrive :size="12" class="shrink-0" />
          <span>Volume</span>
          <span class="ml-auto font-mono">{{ data.volume.size || data.size }}</span>
        </div>
      </PopoverTrigger>
      <PopoverContent class="w-56" :side-offset="8">
        <div class="space-y-2 text-sm">
          <div class="font-semibold text-foreground">{{ data.volume.name }}</div>
          <div class="space-y-1.5">
            <div class="flex justify-between text-muted-foreground">
              <span>Provisioned</span>
              <span class="font-mono">{{ data.volume.size }}</span>
            </div>
            <div class="flex justify-between text-muted-foreground">
              <span>Requested</span>
              <span class="font-mono">{{ data.volume.requestedSize }}</span>
            </div>
          </div>
        </div>
      </PopoverContent>
    </Popover>

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
</style>
