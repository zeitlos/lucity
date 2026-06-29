<script setup lang="ts">
import { computed } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { HardDrive, Plug, Unplug, Server } from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
import type { Service } from '@/composables/useEnvironment';

const UnmountVolumeDocument = graphql(`
  mutation UnmountVolume($volume: VolumeID!) {
    unmountVolume(volume: $volume) {
      id
      mount {
        service
        path
      }
    }
  }
`);

const props = defineProps<{
  volumeId: string;
  volume: {
    name: string;
    size: string;
    mount?: { service: string; path: string } | null;
  };
  services: Service[];
}>();

const emit = defineEmits<{
  (e: 'mount'): void;
  (e: 'unmounted'): void;
}>();

const mountedServiceName = computed(() => {
  if (!props.volume.mount) return null;
  return props.services.find(s => s.id === props.volume.mount!.service)?.name ?? 'Unknown service';
});

const { mutate: unmountVolume, loading: unmounting } = useMutation(UnmountVolumeDocument);

async function handleUnmount() {
  try {
    await unmountVolume({ volume: props.volumeId });
    toast.success('Volume unmounted');
    emit('unmounted');
  } catch (e: unknown) {
    errorToast('Failed to unmount volume', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Storage -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Storage
      </h3>
      <div class="overflow-hidden rounded-lg border">
        <div class="flex items-center gap-3 px-4 py-3">
          <HardDrive :size="16" class="shrink-0 text-muted-foreground" />
          <span class="flex-1 text-sm text-muted-foreground">Size</span>
          <span class="font-mono text-sm font-medium text-foreground">{{ volume.size }}</span>
        </div>
      </div>
    </section>

    <!-- Mount -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Mount
      </h3>

      <!-- Mounted -->
      <div v-if="volume.mount" class="overflow-hidden rounded-lg border">
        <div class="flex items-center gap-3 border-b px-4 py-3">
          <Server :size="16" class="shrink-0 text-muted-foreground" />
          <span class="flex-1 text-sm text-muted-foreground">Service</span>
          <span class="text-sm font-medium text-foreground">{{ mountedServiceName }}</span>
        </div>
        <div class="flex items-center gap-3 px-4 py-3">
          <span class="flex-1 text-sm text-muted-foreground">Mount path</span>
          <span class="font-mono text-sm font-medium text-foreground">{{ volume.mount.path }}</span>
        </div>
      </div>
      <div v-if="volume.mount" class="flex items-center justify-between gap-3 px-1">
        <p class="text-xs text-muted-foreground">
          Unmounting detaches the volume and restarts the service.
        </p>
        <Button
          variant="outline"
          size="sm"
          :disabled="unmounting"
          @click="handleUnmount"
        >
          <Unplug :size="14" class="mr-1" />
          {{ unmounting ? 'Unmounting...' : 'Unmount' }}
        </Button>
      </div>

      <!-- Unmounted -->
      <div
        v-else
        class="flex items-center justify-between gap-3 rounded-lg border border-dashed px-4 py-4"
      >
        <div>
          <p class="text-sm font-medium text-foreground">Not mounted</p>
          <p class="text-xs text-muted-foreground">
            Mount this volume to a service to give it persistent storage.
          </p>
        </div>
        <Button size="sm" @click="emit('mount')">
          <Plug :size="14" class="mr-1" />
          Mount
        </Button>
      </div>
    </section>
  </div>
</template>
