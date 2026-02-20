<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useQuery, useMutation, useLazyQuery } from '@vue/apollo-composable';
import { Loader2, Scan } from 'lucide-vue-next';
import { ProjectQuery } from '@/graphql/projects';
import { DetectServicesQuery, AddServiceMutation } from '@/graphql/services';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from '@/components/ui/sonner';
import { Button } from '@/components/ui/button';
import ServiceCanvas from '@/components/canvas/ServiceCanvas.vue';
import ServicePanel from '@/components/panel/ServicePanel.vue';
import FrameworkIcon from '@/components/FrameworkIcon.vue';
import EmptyState from '@/components/EmptyState.vue';
import CreateCommandPalette from '@/components/CreateCommandPalette.vue';
import { useEnvironment } from '@/composables/useEnvironment';
import { usePanel } from '@/composables/usePanel';
import { errorMessage } from '@/lib/utils';

const route = useRoute();
const projectId = computed(() => route.params.id as string);

const { result, loading, error, refetch } = useQuery(ProjectQuery, () => ({
  id: projectId.value,
}));

const project = computed(() => result.value?.project);

// Environment management
const { setEnvironments, refreshActiveEnvironment } = useEnvironment();
const { isOpen, currentPanel, closePanel } = usePanel();

watch(
  () => project.value?.environments,
  (envs) => {
    if (envs) {
      setEnvironments(envs);
    }
  },
  { immediate: true },
);

// Refresh env data when project refetches
watch(
  () => result.value?.project?.environments,
  (envs) => {
    if (envs) refreshActiveEnvironment(envs);
  },
);

// Selected service for the panel
const selectedService = computed(() => {
  if (!currentPanel.value || !project.value) return null;
  return project.value.services.find(
    (s: { name: string }) => s.name === currentPanel.value!.id
  ) ?? null;
});

// Command palette
const paletteOpen = ref(false);

// Detection
const {
  load: loadDetection,
  result: detectResult,
  loading: detecting,
  error: detectError,
} = useLazyQuery(DetectServicesQuery, () => ({
  projectId: projectId.value,
}), {
  fetchPolicy: 'network-only',
});

const detectedServices = computed(() => detectResult.value?.detectServices ?? []);
const showDetectionBanner = ref(false);

// Auto-detect when project has no services
watch(
  () => project.value,
  (proj) => {
    if (proj && proj.services.length === 0 && !detectResult.value) {
      showDetectionBanner.value = true;
      loadDetection();
    }
  },
  { immediate: true },
);

// Add service
const { mutate: addServiceMutate, loading: addingService } = useMutation(AddServiceMutation);

async function confirmDetectedService(detected: {
  name: string;
  framework: string;
  suggestedPort: number;
}) {
  try {
    const res = await addServiceMutate({
      input: {
        projectId: projectId.value,
        name: detected.name,
        port: detected.suggestedPort,
        public: true,
        framework: detected.framework || undefined,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to add service', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Service added', { description: `${detected.name} configured` });
    showDetectionBanner.value = false;
    refetch();
  } catch (e: unknown) {
    toast.error('Failed to add service', { description: errorMessage(e) });
  }
}

function handleServiceRemoved() {
  closePanel();
  refetch();
}

function handleCreateFromPalette() {
  refetch();
}
</script>

<template>
  <div class="flex h-[calc(100vh-52px)] flex-col">
    <!-- Loading -->
    <div v-if="loading" class="flex flex-1 items-center justify-center">
      <div class="space-y-4 text-center">
        <Skeleton class="mx-auto h-8 w-48" />
        <Skeleton class="mx-auto h-4 w-64" />
      </div>
    </div>

    <!-- Error -->
    <div
      v-else-if="error"
      class="flex flex-1 items-center justify-center p-8"
    >
      <div class="rounded-lg border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">
        Failed to load project: {{ error.message }}
      </div>
    </div>

    <template v-else-if="project">
      <!-- Detection banner (shows above canvas when no services) -->
      <div
        v-if="showDetectionBanner && project.services.length === 0"
        class="shrink-0 border-b bg-card px-4 py-3"
      >
        <div v-if="detecting" class="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 :size="16" class="animate-spin" />
          Scanning repository for services...
        </div>

        <div
          v-else-if="detectError"
          class="flex items-center justify-between"
        >
          <span class="text-sm text-destructive">Detection failed: {{ detectError.message }}</span>
          <Button
            variant="outline"
            size="sm"
            @click="loadDetection()"
          >
            Retry
          </Button>
        </div>

        <div
          v-else-if="detectedServices.length > 0"
          class="space-y-2"
        >
          <div
            v-for="detected in detectedServices"
            :key="detected.name"
            class="flex items-center justify-between"
          >
            <div class="flex items-center gap-2">
              <FrameworkIcon :framework="detected.framework" :size="20" />
              <span class="text-sm">
                {{ detected.framework || detected.provider }} app detected
                <span class="text-muted-foreground">(port {{ detected.suggestedPort }})</span>
              </span>
            </div>
            <Button
              size="sm"
              :disabled="addingService"
              @click="confirmDetectedService(detected)"
            >
              {{ addingService ? 'Adding...' : 'Confirm & Add' }}
            </Button>
          </div>
        </div>
      </div>

      <!-- Canvas + Panel split -->
      <div class="flex flex-1 overflow-hidden">
        <!-- Canvas -->
        <div
          :class="[
            'transition-all duration-300 ease-in-out',
            isOpen ? 'w-[45%]' : 'w-full',
          ]"
        >
          <template v-if="project.services.length > 0">
            <ServiceCanvas
              :services="project.services"
              @create="paletteOpen = true"
            />
          </template>
          <template v-else>
            <div class="flex h-full items-center justify-center">
              <EmptyState
                title="No services yet"
                description="Detect services from your repository or create one manually."
                pattern="crosshatch"
              >
                <template #action>
                  <div class="flex gap-2">
                    <Button
                      v-if="!showDetectionBanner"
                      variant="outline"
                      @click="showDetectionBanner = true; loadDetection();"
                    >
                      <Scan :size="14" class="mr-2" />
                      Detect Services
                    </Button>
                    <Button @click="paletteOpen = true">
                      Create Service
                    </Button>
                  </div>
                </template>
              </EmptyState>
            </div>
          </template>
        </div>

        <!-- Service Detail Panel -->
        <Transition name="slide-panel">
          <div
            v-if="isOpen && selectedService"
            class="w-[55%] shrink-0"
          >
            <ServicePanel
              :project-id="projectId"
              :service="selectedService"
              @close="closePanel"
              @service-removed="handleServiceRemoved"
            />
          </div>
        </Transition>
      </div>
    </template>

    <!-- Command Palette -->
    <CreateCommandPalette
      v-model:open="paletteOpen"
      context="project"
      :project-id="projectId"
      @created="handleCreateFromPalette"
    />
  </div>
</template>

<style scoped>
.slide-panel-enter-active {
  transition: transform 0.3s ease, opacity 0.2s ease;
}

.slide-panel-leave-active {
  transition: transform 0.2s ease, opacity 0.15s ease;
}

.slide-panel-enter-from,
.slide-panel-leave-to {
  transform: translateX(100%);
  opacity: 0;
}
</style>
