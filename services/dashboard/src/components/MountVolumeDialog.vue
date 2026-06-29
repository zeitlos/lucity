<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Server, Plug, ArrowLeft, Check } from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
import type { Service } from '@/composables/useEnvironment';

const MountVolumeDocument = graphql(`
  mutation MountVolume($volume: VolumeID!, $service: ServiceID!, $path: String!) {
    mountVolume(volume: $volume, service: $service, path: $path) {
      id
      mount {
        service
        path
      }
    }
  }
`);

const props = defineProps<{
  open: boolean;
  volumeId: string | null;
  volumeName: string;
  services: Service[];
  mountedServiceIds: string[];
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
  (e: 'mounted'): void;
}>();

const step = ref<'service' | 'path'>('service');
const selectedServiceId = ref<string>('');
const path = ref('/data');

watch(() => props.open, (open) => {
  if (open) {
    step.value = 'service';
    selectedServiceId.value = '';
    path.value = '/data';
  }
});

function hasVolume(serviceId: string): boolean {
  return props.mountedServiceIds.includes(serviceId);
}

const selectedServiceName = computed(
  () => props.services.find(s => s.id === selectedServiceId.value)?.name ?? '',
);

const pathError = computed(() => {
  const p = path.value.trim();
  if (!p) return 'Enter a mount path.';
  if (!p.startsWith('/')) return 'Path must be absolute (start with /).';
  if (p === '/') return 'Path cannot be the root directory.';
  if (p.includes('..')) return 'Path cannot contain "..".';
  if (/\s/.test(p)) return 'Path cannot contain spaces.';
  return null;
});

const { mutate: mountVolume, loading: mounting } = useMutation(MountVolumeDocument);

function close() {
  emit('update:open', false);
}

function selectService(serviceId: string) {
  if (hasVolume(serviceId)) return;
  selectedServiceId.value = serviceId;
}

function goToPath() {
  if (!selectedServiceId.value) return;
  step.value = 'path';
}

async function handleMount() {
  if (!props.volumeId || !selectedServiceId.value || pathError.value) return;

  try {
    await mountVolume({
      volume: props.volumeId,
      service: selectedServiceId.value,
      path: path.value.trim(),
    });
    toast.success('Volume mounted', { description: `to ${selectedServiceName.value}` });
    close();
    emit('mounted');
  } catch (e: unknown) {
    errorToast('Failed to mount volume', { description: errorMessage(e) });
  }
}
</script>

<template>
  <Dialog
    :open="open"
    @update:open="emit('update:open', $event)"
  >
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>
          Mount <span class="font-mono">{{ volumeName }}</span>
        </DialogTitle>
        <DialogDescription>
          {{ step === 'service'
            ? 'Choose the service that should read and write this volume.'
            : 'Choose where the volume is mounted inside the service.' }}
        </DialogDescription>
      </DialogHeader>

      <!-- Step 1: select service -->
      <div v-if="step === 'service'" class="space-y-1.5">
        <p
          v-if="services.length === 0"
          class="rounded-lg border border-dashed px-4 py-6 text-center text-sm text-muted-foreground"
        >
          Create a service first, then mount this volume to it.
        </p>
        <button
          v-for="service in services"
          :key="service.id"
          type="button"
          :disabled="hasVolume(service.id)"
          class="flex w-full items-center gap-3 rounded-lg border px-3 py-2.5 text-left text-sm transition-colors disabled:cursor-not-allowed disabled:opacity-50"
          :class="selectedServiceId === service.id ? 'border-primary bg-primary/5' : 'border-border hover:bg-accent'"
          @click="selectService(service.id)"
        >
          <Server :size="16" class="shrink-0 text-muted-foreground" />
          <span class="flex-1 truncate font-medium text-foreground">{{ service.name }}</span>
          <span
            v-if="hasVolume(service.id)"
            class="text-xs text-muted-foreground"
          >already has a volume</span>
          <Check
            v-else-if="selectedServiceId === service.id"
            :size="16"
            class="shrink-0 text-primary"
          />
        </button>
      </div>

      <!-- Step 2: mount path -->
      <div v-else class="space-y-4">
        <div class="flex items-center gap-3 rounded-lg border px-3 py-2.5 text-sm">
          <Server :size="16" class="shrink-0 text-muted-foreground" />
          <span class="flex-1 truncate font-medium text-foreground">{{ selectedServiceName }}</span>
        </div>
        <div class="space-y-2">
          <Label for="mount-path">Mount path</Label>
          <Input
            id="mount-path"
            v-model="path"
            placeholder="/data"
            spellcheck="false"
            class="font-mono"
            @keydown.enter.prevent="handleMount"
          />
          <p v-if="pathError" class="text-xs text-destructive">{{ pathError }}</p>
          <p v-else class="text-xs text-muted-foreground">
            Mounting attaches the volume and restarts the service.
          </p>
        </div>
      </div>

      <DialogFooter>
        <Button
          v-if="step === 'path'"
          variant="ghost"
          class="mr-auto"
          @click="step = 'service'"
        >
          <ArrowLeft :size="14" class="mr-1" />
          Back
        </Button>
        <Button
          v-if="step === 'service'"
          :disabled="!selectedServiceId"
          @click="goToPath"
        >
          Next
        </Button>
        <Button
          v-else
          :disabled="mounting || !!pathError"
          @click="handleMount"
        >
          <Plug :size="14" class="mr-1" />
          {{ mounting ? 'Mounting...' : 'Mount' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
