<script setup lang="ts">
import { computed } from 'vue';
import { X, HardDrive, ArrowLeft } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';
import type { VolumeInfo } from '@/composables/useEnvironment';

const props = defineProps<{
  volume: VolumeInfo;
  databaseName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'back'): void;
}>();

const usagePercent = computed(() => {
  if (props.volume.capacityBytes <= 0) return 0;
  return Math.min(100, Math.round((props.volume.usedBytes / props.volume.capacityBytes) * 100));
});

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

const usedFormatted = computed(() => formatBytes(props.volume.usedBytes));
const capacityFormatted = computed(() => formatBytes(props.volume.capacityBytes));

const barColor = computed(() => {
  if (usagePercent.value >= 90) return 'bg-destructive';
  if (usagePercent.value >= 75) return 'bg-orange-500';
  return 'bg-primary';
});
</script>

<template>
  <div class="flex h-full flex-col rounded-lg border bg-card/80 shadow-sm backdrop-blur-sm [background-image:var(--gradient-card)]">
    <!-- Header -->
    <div class="flex shrink-0 items-center justify-between border-b px-4 py-3">
      <div class="flex items-center gap-3">
        <Button
          variant="ghost"
          size="icon"
          class="h-7 w-7"
          @click="emit('back')"
        >
          <ArrowLeft :size="16" />
        </Button>
        <HardDrive :size="20" class="shrink-0 text-muted-foreground" />
        <h2 class="text-lg font-semibold text-foreground">Volume</h2>
      </div>

      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        @click="emit('close')"
      >
        <X :size="16" />
      </Button>
    </div>

    <!-- Content -->
    <div class="flex-1 space-y-6 px-4 py-4">
      <!-- Usage bar -->
      <div class="space-y-2">
        <div class="flex items-center justify-between text-sm">
          <span class="text-muted-foreground">Disk Usage</span>
          <span class="font-mono font-medium text-foreground">{{ usagePercent }}%</span>
        </div>
        <div class="h-2.5 w-full overflow-hidden rounded-full bg-muted">
          <div
            :class="['h-full rounded-full transition-all duration-500', barColor]"
            :style="{ width: usagePercent + '%' }"
          />
        </div>
        <div
          v-if="volume.capacityBytes > 0"
          class="flex justify-between text-xs text-muted-foreground"
        >
          <span>{{ usedFormatted }} used</span>
          <span>{{ capacityFormatted }} total</span>
        </div>
      </div>

      <!-- Details -->
      <div class="space-y-3">
        <h3 class="text-sm font-medium text-foreground">Details</h3>
        <div class="space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-muted-foreground">Name</span>
            <span class="truncate pl-4 font-mono text-foreground">{{ volume.name }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-muted-foreground">Database</span>
            <span class="font-mono text-foreground">{{ databaseName }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-muted-foreground">Provisioned</span>
            <span class="font-mono text-foreground">{{ volume.size }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-muted-foreground">Requested</span>
            <span class="font-mono text-foreground">{{ volume.requestedSize }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
