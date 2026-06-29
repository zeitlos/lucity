<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { onKeyStroke } from '@vueuse/core';
import { Server, Plug, ArrowLeft, Search, X } from '@lucide/vue';
import { graphql } from '@/gql';
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
const search = ref('');
const focusedIndex = ref(0);
const inputRef = ref<HTMLInputElement>();

watch(() => props.open, (open) => {
  if (open) {
    step.value = 'service';
    selectedServiceId.value = '';
    path.value = '/data';
    search.value = '';
    focusedIndex.value = 0;
    nextTick(() => inputRef.value?.focus());
  }
});

watch(step, () => {
  nextTick(() => inputRef.value?.focus());
});

watch(search, () => {
  focusedIndex.value = 0;
});

function hasVolume(serviceId: string): boolean {
  return props.mountedServiceIds.includes(serviceId);
}

const filteredServices = computed(() => {
  if (!search.value) return props.services;
  const q = search.value.toLowerCase();
  return props.services.filter(s => s.name.toLowerCase().includes(q));
});

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
  step.value = 'path';
}

function scrollFocusedIntoView() {
  document.querySelector('[data-focused="true"]')?.scrollIntoView({ block: 'nearest' });
}

onKeyStroke('Escape', () => {
  if (!props.open) return;
  if (step.value === 'path') {
    step.value = 'service';
  } else {
    close();
  }
});

onKeyStroke('ArrowDown', (e) => {
  if (!props.open || step.value !== 'service') return;
  if (filteredServices.value.length === 0) return;
  e.preventDefault();
  focusedIndex.value = (focusedIndex.value + 1) % filteredServices.value.length;
  nextTick(scrollFocusedIntoView);
});

onKeyStroke('ArrowUp', (e) => {
  if (!props.open || step.value !== 'service') return;
  if (filteredServices.value.length === 0) return;
  e.preventDefault();
  focusedIndex.value = (focusedIndex.value - 1 + filteredServices.value.length) % filteredServices.value.length;
  nextTick(scrollFocusedIntoView);
});

onKeyStroke('Enter', (e) => {
  if (!props.open) return;
  if (step.value === 'service') {
    const svc = filteredServices.value[focusedIndex.value];
    if (svc && !hasVolume(svc.id)) {
      e.preventDefault();
      selectService(svc.id);
    }
  } else if (!mounting.value && !pathError.value) {
    e.preventDefault();
    handleMount();
  }
});

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
  <Teleport to="body">
    <Transition name="palette">
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]"
      >
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-background/80 backdrop-blur-sm"
          @click="close"
        />

        <!-- Palette -->
        <div class="relative z-10 w-full max-w-lg rounded-xl border bg-popover shadow-2xl">
          <!-- Step: select service -->
          <template v-if="step === 'service'">
            <div class="flex items-center border-b px-3">
              <Search :size="18" class="shrink-0 text-muted-foreground" />
              <input
                ref="inputRef"
                v-model="search"
                :placeholder="`Mount ${volumeName} to which service?`"
                class="flex h-12 w-full bg-transparent px-3 text-sm outline-none placeholder:text-muted-foreground"
              />
              <button
                class="shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="close"
              >
                <X :size="16" />
              </button>
            </div>

            <div class="max-h-[320px] overflow-y-auto p-1">
              <p class="px-2 py-1.5 text-xs font-medium text-muted-foreground">Services</p>
              <button
                v-for="(service, index) in filteredServices"
                :key="service.id"
                :data-focused="focusedIndex === index"
                :disabled="hasVolume(service.id)"
                class="flex w-full items-center gap-2 rounded-lg px-2 py-2.5 text-sm text-popover-foreground transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                :class="focusedIndex === index ? 'bg-accent' : 'hover:bg-accent'"
                @click="selectService(service.id)"
                @mouseenter="focusedIndex = index"
              >
                <Server :size="16" class="shrink-0 text-muted-foreground" />
                <span class="flex-1 truncate text-left font-medium">{{ service.name }}</span>
                <span
                  v-if="hasVolume(service.id)"
                  class="shrink-0 text-xs text-muted-foreground"
                >already has a volume</span>
              </button>
              <p
                v-if="filteredServices.length === 0"
                class="px-2 py-6 text-center text-sm text-muted-foreground"
              >
                {{ services.length === 0 ? 'Create a service first, then mount this volume to it.' : 'No services found.' }}
              </p>
            </div>
          </template>

          <!-- Step: mount path -->
          <template v-else>
            <div class="flex h-12 items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="step = 'service'"
              >
                <ArrowLeft :size="16" />
              </button>
              <span class="ml-1 text-sm font-medium text-foreground">Mount path</span>
              <div class="flex-1" />
              <button
                class="shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="close"
              >
                <X :size="16" />
              </button>
            </div>
            <div class="space-y-4 p-4">
              <div class="flex items-center gap-3 rounded-md border px-3 py-2.5 text-sm">
                <Server :size="16" class="shrink-0 text-muted-foreground" />
                <span class="flex-1 truncate font-medium text-foreground">{{ selectedServiceName }}</span>
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium text-foreground">Mount path</label>
                <input
                  ref="inputRef"
                  v-model="path"
                  spellcheck="false"
                  placeholder="/data"
                  class="flex h-9 w-full rounded-md border bg-transparent px-3 py-1 font-mono text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
                <p v-if="pathError" class="text-xs text-destructive">{{ pathError }}</p>
                <p v-else class="text-xs text-muted-foreground">
                  Mounting attaches the volume and restarts the service.
                </p>
              </div>
              <button
                class="inline-flex h-9 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                :disabled="mounting || !!pathError"
                @click="handleMount"
              >
                <Plug :size="14" class="mr-1.5" />
                {{ mounting ? 'Mounting...' : 'Mount Volume' }}
              </button>
            </div>
          </template>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.palette-enter-active,
.palette-leave-active {
  transition: opacity 0.15s ease;
}

.palette-enter-active .relative,
.palette-leave-active .relative {
  transition: transform 0.15s ease, opacity 0.15s ease;
}

.palette-enter-from,
.palette-leave-to {
  opacity: 0;
}

.palette-enter-from .relative,
.palette-leave-to .relative {
  transform: scale(0.96);
  opacity: 0;
}
</style>
