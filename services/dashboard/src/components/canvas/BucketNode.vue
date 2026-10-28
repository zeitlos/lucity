<script setup lang="ts">
import { bucketStatus } from '@/lib/status';
import { computed } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import BucketIcon from '@/components/BucketIcon.vue';
import { BucketStatus } from '@/gql/graphql';
import { Status } from '@/components/ui/status';

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

const status = computed(() => bucketStatus(props.data.status));
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

      <!-- Status row -->
      <div class="mt-4 flex items-center justify-between border-t border-border/50 pt-4">
        <Status :tone="status.tone" size="sm">{{ status.label }}</Status>
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
