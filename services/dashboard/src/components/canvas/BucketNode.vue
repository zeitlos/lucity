<script setup lang="ts">
import { computed } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import BucketIcon from '@/components/BucketIcon.vue';
import { BucketStatus } from '@/gql/graphql';
import { Status } from '@/components/ui/status';
import { formatBytes } from '@/lib/utils';

const props = defineProps<{
  data: {
    name: string;
    region: string;
    sizeBytes: number;
    status: BucketStatus;
  };
  selected?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select'): void;
}>();

const statusTone = computed(() => {
  switch (props.data.status) {
    case BucketStatus.Ready:
      return 'ok' as const;
    case BucketStatus.Failed:
      return 'danger' as const;
    case BucketStatus.Pending:
      return 'progress' as const;
    default:
      return 'neutral' as const;
  }
});

const statusLabel = computed(() => {
  switch (props.data.status) {
    case BucketStatus.Ready:
      return 'Ready';
    case BucketStatus.Pending:
      return 'Provisioning';
    case BucketStatus.Failed:
      return 'Failed';
    default:
      return 'Unknown';
  }
});
</script>

<template>
  <div class="bucket-node-wrapper">
    <div
      :class="[
        'bucket-node group cursor-pointer rounded-xl border px-6 py-5 shadow-sm transition-all duration-200',
        'hover:shadow-md',
        selected ? 'border-primary shadow-md' : 'border-border',
      ]"
      style="width: 280px;"
      @click="emit('select')"
    >
      <!-- Header: icon + name -->
      <div class="flex items-center gap-3">
        <BucketIcon :size="28" class="shrink-0 text-muted-foreground" />
        <span class="truncate font-semibold text-foreground">{{ data.name }}</span>
      </div>

      <!-- Region -->
      <div class="mt-3">
        <div class="flex items-center gap-1.5 text-xs text-muted-foreground">
          <span class="font-mono">Object Storage &middot; {{ data.region }}</span>
        </div>
      </div>

      <!-- Status row -->
      <div class="mt-4 flex items-center justify-between border-t border-border/50 pt-4">
        <Status :tone="statusTone" class="text-[0.65rem]">{{ statusLabel }}</Status>
        <span class="text-[0.65rem] font-mono text-muted-foreground">{{ formatBytes(data.sizeBytes) }}</span>
      </div>
    </div>

    <!-- Vue Flow handles -->
    <Handle type="source" :position="Position.Bottom" class="!invisible" />
    <Handle type="target" :position="Position.Top" class="!invisible" />
  </div>
</template>

<style scoped>
.bucket-node-wrapper {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.bucket-node {
  position: relative;
  z-index: 1;
  background: linear-gradient(
    to bottom,
    var(--card) 0%,
    color-mix(in oklch, var(--card) 94%, var(--muted)) 100%
  );
}
</style>
