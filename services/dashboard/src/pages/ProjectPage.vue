<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { ProjectQuery } from '@/graphql/projects';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import ServiceCanvas from '@/components/canvas/ServiceCanvas.vue';
import ServicePanel from '@/components/panel/ServicePanel.vue';
import EmptyState from '@/components/EmptyState.vue';
import CreateCommandPalette from '@/components/CreateCommandPalette.vue';
import { useEnvironment } from '@/composables/useEnvironment';
import { usePanel } from '@/composables/usePanel';

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

function handleServiceRemoved() {
  closePanel();
  refetch();
}

function handleCreateFromPalette() {
  refetch();
}
</script>

<template>
  <div class="flex h-[calc(100vh-52px-0.75rem)] flex-col">
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
      <!-- Canvas + Panel overlay -->
      <div class="relative flex-1 overflow-hidden p-3">
        <!-- Canvas (always full width) -->
        <div class="h-full w-full overflow-hidden rounded-lg border bg-card/80 shadow-sm backdrop-blur-sm [background-image:var(--gradient-card)]">
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
                description="Create a service manually to get started."
                pattern="crosshatch"
              >
                <template #action>
                  <Button @click="paletteOpen = true">
                    Create Service
                  </Button>
                </template>
              </EmptyState>
            </div>
          </template>
        </div>

        <!-- Service Detail Panel (overlays from right) -->
        <Transition name="slide-panel">
          <div
            v-if="isOpen && selectedService"
            class="absolute inset-y-3 right-3 w-[55%] shadow-xl"
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
