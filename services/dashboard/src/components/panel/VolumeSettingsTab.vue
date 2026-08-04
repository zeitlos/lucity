<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Trash2, HardDrive, Server, Plug, Unplug } from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import { Slider } from '@/components/ui/slider';
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
import type { Service } from '@/composables/useEnvironment';

const DeleteVolumeDocument = graphql(`
  mutation DeleteVolume($volume: VolumeID!) {
    deleteVolume(volume: $volume)
  }
`);

const ExpandVolumeDocument = graphql(`
  mutation ExpandVolume($volume: VolumeID!, $size: String!) {
    expandVolume(volume: $volume, size: $size) {
      id
      size
    }
  }
`);

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
  (e: 'volume-removed'): void;
  (e: 'mount'): void;
  (e: 'unmounted'): void;
  (e: 'expanded'): void;
}>();

const mountedServiceName = computed(() => {
  if (!props.volume.mount) return null;
  return props.services.find(s => s.id === props.volume.mount!.service)?.name ?? 'Unknown service';
});

const volumeSizes = ['10Gi', '16Gi', '32Gi', '64Gi', '128Gi', '256Gi', '512Gi', '1Ti'];

const parseGib = (size: string) => {
  const match = /^(\d+)(Gi|Ti)$/.exec(size);
  return match ? Number(match[1]) * (match[2] === 'Ti' ? 1024 : 1) : 0;
};

const currentSize = computed(() => props.volume.size);

const minIndex = computed(() => {
  const index = volumeSizes.findIndex(size => parseGib(size) >= parseGib(currentSize.value));
  return index === -1 ? volumeSizes.length - 1 : index;
});

const selectedSize = ref(props.volume.size);

watch(currentSize, () => { selectedSize.value = volumeSizes[minIndex.value]!; }, { immediate: true });

const storageChanged = computed(() => parseGib(selectedSize.value) > parseGib(currentSize.value));
const storageSaving = ref(false);

const { mutate: expandVolume } = useMutation(ExpandVolumeDocument);

async function handleExpand() {
  storageSaving.value = true;
  try {
    const res = await expandVolume({ volume: props.volumeId, size: selectedSize.value });

    if (res?.errors?.length) {
      errorToast('Failed to expand volume', { description: res.errors.map(e => e.message).join(', ') });
      return;
    }

    toast.success('Volume expanded');
    emit('expanded');
  } catch (e: unknown) {
    errorToast('Failed to expand volume', { description: errorMessage(e) });
  } finally {
    storageSaving.value = false;
  }
}

const { mutate: unmountVolume, loading: unmounting } = useMutation(UnmountVolumeDocument);
const unmountDialogOpen = ref(false);

async function handleUnmount() {
  try {
    await unmountVolume({ volume: props.volumeId });
    toast.success('Volume unmounted');
    unmountDialogOpen.value = false;
    emit('unmounted');
  } catch (e: unknown) {
    errorToast('Failed to unmount volume', { description: errorMessage(e) });
  }
}

const { mutate: deleteVolume, loading: deleting } = useMutation(DeleteVolumeDocument);
const deleteDialogOpen = ref(false);

async function handleDelete() {
  try {
    await deleteVolume({ volume: props.volumeId });
    toast.success('Volume deleted');
    deleteDialogOpen.value = false;
    emit('volume-removed');
  } catch (e: unknown) {
    errorToast('Failed to delete volume', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="space-y-6">
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
        <AlertDialog v-model:open="unmountDialogOpen">
          <AlertDialogTrigger as-child>
            <Button
              variant="outline"
              size="sm"
              :disabled="unmounting"
            >
              <Unplug :size="14" class="mr-1" />
              {{ unmounting ? 'Unmounting...' : 'Unmount' }}
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Unmount "{{ volume.name }}" from {{ mountedServiceName }}?</AlertDialogTitle>
              <AlertDialogDescription>
                This detaches the volume and restarts the service. Your data is kept,
                and you can remount the volume later.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel :disabled="unmounting">Cancel</AlertDialogCancel>
              <Button
                :disabled="unmounting"
                @click="handleUnmount"
              >
                {{ unmounting ? 'Unmounting...' : 'Unmount' }}
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
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

    <!-- Storage -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Storage
      </h3>

      <div class="overflow-hidden rounded-lg border">
        <div class="divide-y">
          <div class="space-y-3 px-4 py-3">
            <div class="flex items-center gap-3">
              <HardDrive :size="16" class="shrink-0 text-muted-foreground" />
              <span class="flex-1 text-sm text-muted-foreground">Size</span>
              <span class="font-mono text-sm font-semibold text-foreground">{{ selectedSize }}</span>
            </div>
            <Slider
              :model-value="[volumeSizes.indexOf(selectedSize)]"
              :min="minIndex"
              :max="volumeSizes.length - 1"
              :step="1"
              @update:model-value="selectedSize = volumeSizes[$event?.[0] ?? minIndex]!"
            />
            <div class="flex justify-between text-[10px] text-muted-foreground">
              <span>{{ volume.size }}</span>
              <span>{{ volumeSizes[volumeSizes.length - 1] }}</span>
            </div>
          </div>
          <div class="flex items-center justify-between gap-3 px-4 py-3">
            <p class="text-xs text-muted-foreground">
              You can add more space anytime, but storage can't be made smaller later.
            </p>
            <Button
              size="sm"
              :disabled="!storageChanged || storageSaving"
              @click="handleExpand"
            >
              {{ storageSaving ? 'Saving...' : 'Save' }}
            </Button>
          </div>
        </div>
      </div>
    </section>

    <!-- Danger Zone -->
    <section class="mt-8">
      <div class="relative overflow-hidden rounded-lg border border-destructive/20">
        <div class="pattern-crosshatch pointer-events-none absolute inset-0 opacity-[0.04]" />
        <div class="relative border-b border-destructive/15 bg-destructive/[0.03] px-4 py-2.5">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-destructive/70">
            Danger Zone
          </h3>
        </div>
        <div class="relative px-4 py-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-foreground">Delete Volume</p>
              <p class="text-xs text-muted-foreground">
                Permanently delete this volume and all its data.
              </p>
            </div>
            <AlertDialog v-model:open="deleteDialogOpen">
              <AlertDialogTrigger as-child>
                <Button variant="destructive" size="sm">
                  <Trash2 :size="14" class="mr-1" />
                  Delete
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete volume "{{ volume.name }}"?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This will permanently delete the volume and all the data stored on it.
                    <template v-if="volume.mount">
                      The volume is still mounted; unmount it first if you only want to detach it.
                    </template>
                    This action cannot be undone.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel :disabled="deleting">Cancel</AlertDialogCancel>
                  <Button
                    variant="destructive"
                    :disabled="deleting"
                    @click="handleDelete"
                  >
                    {{ deleting ? 'Deleting...' : 'Delete Volume' }}
                  </Button>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>
