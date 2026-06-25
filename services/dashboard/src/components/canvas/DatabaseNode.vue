<script setup lang="ts">
import { computed } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import { DatabaseStatus } from '@/gql/graphql';
import { Status } from '@/components/ui/status';

const props = defineProps<{
  data: {
    name: string;
    version: string;
    instances: number;
    size: string;
    status: DatabaseStatus;
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
}>();

const statusTone = computed(() => {
  switch (props.data.status) {
    case DatabaseStatus.Healthy:
      return 'ok' as const;
    case DatabaseStatus.Failed:
      return 'danger' as const;
    case DatabaseStatus.Pending:
    case DatabaseStatus.Updating:
      return 'warn' as const;
    default:
      return 'neutral' as const;
  }
});

const statusLabel = computed(() => {
  switch (props.data.status) {
    case DatabaseStatus.Healthy:
      return 'Online';
    case DatabaseStatus.Degraded:
      return 'Degraded';
    case DatabaseStatus.Updating:
      return 'Updating';
    case DatabaseStatus.Pending:
      return 'Provisioning';
    case DatabaseStatus.Failed:
      return 'Failed';
    case DatabaseStatus.Stopped:
      return 'Stopped';
    default:
      return 'Unknown';
  }
});
</script>

<template>
  <div class="database-node-wrapper">
    <div
      :class="[
        'database-node group cursor-pointer rounded-xl border px-6 py-5 shadow-sm transition-all duration-200',
        'hover:shadow-md',
        selected ? 'border-primary shadow-md' : 'border-border',
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
        <Status :tone="statusTone" class="text-[0.65rem]">{{ statusLabel }}</Status>
        <span class="text-[0.65rem] font-mono text-muted-foreground">{{ data.size }}</span>
      </div>
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
</style>
