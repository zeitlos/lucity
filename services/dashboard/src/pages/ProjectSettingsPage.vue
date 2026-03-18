<script setup lang="ts">
import { computed, ref, reactive } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { ArrowLeft, Trash2, ChevronDown, ChevronRight } from 'lucide-vue-next';
import { ProjectQuery, DeleteProjectMutation, DeleteEnvironmentMutation } from '@/graphql/projects';
import { EnvironmentResourcesQuery, SetEnvironmentResourcesMutation } from '@/graphql/billing';
import { apolloClient } from '@/lib/apollo';
import { useEnvironment } from '@/composables/useEnvironment';
import SharedVariablesEditor from '@/components/SharedVariablesEditor.vue';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Slider } from '@/components/ui/slider';
import {
  AlertDialog,
  AlertDialogAction,
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

const route = useRoute();
const router = useRouter();
const projectId = computed(() => route.params.id as string);

const { result, loading } = useQuery(ProjectQuery, () => ({
  id: projectId.value,
}));

const project = computed(() => result.value?.project);

// Environment management — reuse global composable so SharedVariablesEditor picks them up
const { setEnvironments, environments, activeEnvironment, setEnvironment } = useEnvironment();

import { watch } from 'vue';
watch(
  () => project.value?.environments,
  (envs) => {
    if (envs) setEnvironments(envs);
  },
  { immediate: true },
);

// Settings sections — driven by route param
const validSections = ['general', 'environments', 'variables'];
const sections = [
  { id: 'general', label: 'General' },
  { id: 'environments', label: 'Environments' },
  { id: 'variables', label: 'Variables' },
];

const activeSection = computed({
  get: () => {
    const s = route.params.section as string | undefined;
    return s && validSections.includes(s) ? s : 'general';
  },
  set: (val: string) => {
    router.replace({
      name: 'project-settings',
      params: { id: projectId.value, section: val === 'general' ? undefined : val },
      query: route.query,
    });
  },
});

// Delete project
const { mutate: deleteProjectMutate, loading: deleting } = useMutation(DeleteProjectMutation);

async function handleDeleteProject() {
  try {
    const res = await deleteProjectMutate({ id: projectId.value });

    if (res?.errors?.length) {
      errorToast('Failed to delete project', {
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
      });
      return;
    }

    apolloClient.cache.evict({ id: `Project:${projectId.value}` });
    apolloClient.cache.gc();

    toast.success('Project deleted');
    router.push({ name: 'projects' });
  } catch (e: unknown) {
    errorToast('Failed to delete project', { description: errorMessage(e) });
  }
}

// Resource presets — slider index maps to value
const cpuSteps = [500, 1000, 2000, 4000, 8000, 16000];
const memorySteps = [512, 1024, 2048, 4096, 8192, 16384];
const diskSteps = [1024, 2048, 5120, 10240, 20480, 51200];

function toIndex(steps: number[], value: number) {
  const idx = steps.indexOf(value);
  return idx >= 0 ? idx : 0;
}

function formatCpu(m: number) {
  return m >= 1000 ? `${m / 1000} vCPU` : `${m / 1000} vCPU`;
}

function formatMB(mb: number) {
  return mb >= 1024 ? `${mb / 1024} GB` : `${mb} MB`;
}

// Environment resources
interface EnvResourceState {
  loading: boolean;
  loaded: boolean;
  saving: boolean;
  tier: string;
  cpuMillicores: number;
  memoryMB: number;
  diskMB: number;
}

const expandedEnv = ref<string | null>(null);
const envResources: Record<string, EnvResourceState> = reactive({});

const { mutate: setResourcesMutate } = useMutation(SetEnvironmentResourcesMutation);

async function toggleEnvExpand(envName: string) {
  if (expandedEnv.value === envName) {
    expandedEnv.value = null;
    return;
  }
  expandedEnv.value = envName;

  if (envResources[envName]?.loaded) return;

  envResources[envName] = {
    loading: true,
    loaded: false,
    saving: false,
    tier: 'ECO',
    cpuMillicores: 1000,
    memoryMB: 1024,
    diskMB: 1024,
  };

  try {
    const { data } = await apolloClient.query({
      query: EnvironmentResourcesQuery,
      variables: { projectId: projectId.value, environment: envName },
      fetchPolicy: 'network-only',
    });
    if (data?.environmentResources) {
      const r = data.environmentResources;
      envResources[envName]!.tier = r.tier;
      envResources[envName]!.cpuMillicores = r.allocation.cpuMillicores;
      envResources[envName]!.memoryMB = r.allocation.memoryMB;
      envResources[envName]!.diskMB = r.allocation.diskMB;
    }
  } catch {
    // No resources set yet — keep defaults
  } finally {
    envResources[envName]!.loading = false;
    envResources[envName]!.loaded = true;
  }
}

// Auto-expand environment from query param
watch(
  () => [route.query.env, project.value?.environments],
  () => {
    const envName = route.query.env as string | undefined;
    if (envName && project.value?.environments?.some((e: { name: string }) => e.name === envName)) {
      if (expandedEnv.value !== envName) {
        toggleEnvExpand(envName);
      }
    }
  },
  { immediate: true },
);

async function handleSaveResources(envName: string) {
  const state = envResources[envName];
  if (!state) return;

  state.saving = true;
  try {
    await setResourcesMutate({
      input: {
        projectId: projectId.value,
        environment: envName,
        tier: state.tier,
        cpuMillicores: state.cpuMillicores,
        memoryMB: state.memoryMB,
        diskMB: state.diskMB,
      },
    });
    toast.success(`Resources updated for "${envName}"`);
  } catch (e: unknown) {
    errorToast('Failed to update resources', { description: errorMessage(e) });
  } finally {
    state.saving = false;
  }
}

// Delete environment
const { mutate: deleteEnvironmentMutate, loading: deletingEnv } = useMutation(DeleteEnvironmentMutation, {
  refetchQueries: () => [{ query: ProjectQuery, variables: { id: projectId.value } }],
});
const envToDelete = ref<string | null>(null);

async function handleDeleteEnvironment() {
  const name = envToDelete.value;
  if (!name) return;

  try {
    const res = await deleteEnvironmentMutate({
      projectId: projectId.value,
      environment: name,
    });

    if (res?.errors?.length) {
      errorToast('Failed to delete environment', {
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
      });
      return;
    }

    // If the deleted environment was active, switch to another one
    if (activeEnvironment.value?.name === name) {
      const remaining = environments.value.filter(e => e.name !== name);
      if (remaining.length > 0) {
        setEnvironment(remaining[0]!);
      }
    }

    toast.success(`Environment "${name}" deleted`);
  } catch (e: unknown) {
    errorToast('Failed to delete environment', { description: errorMessage(e) });
  } finally {
    envToDelete.value = null;
  }
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

    <template v-else-if="project">
      <div class="flex flex-1 overflow-hidden p-3">
        <div class="mx-auto flex w-full max-w-4xl gap-6 overflow-hidden rounded-lg border bg-card/80 shadow-sm backdrop-blur-sm [background-image:var(--gradient-card)]">
          <!-- Sidebar -->
          <nav class="w-48 shrink-0 border-r p-4">
            <div class="mb-4">
              <button
                class="flex items-center gap-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
                @click="router.push({ name: 'project', params: { id: projectId } })"
              >
                <ArrowLeft :size="12" />
                Back to project
              </button>
            </div>
            <h2 class="mb-3 text-sm font-semibold text-foreground">Settings</h2>
            <ul class="space-y-1">
              <li v-for="section in sections" :key="section.id">
                <button
                  class="w-full rounded-md px-3 py-1.5 text-left text-sm transition-colors"
                  :class="activeSection === section.id
                    ? 'bg-accent text-accent-foreground font-medium'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'"
                  @click="activeSection = section.id"
                >
                  {{ section.label }}
                </button>
              </li>
            </ul>
          </nav>

          <!-- Content -->
          <div class="flex-1 overflow-y-auto p-6">
            <!-- General -->
            <div v-if="activeSection === 'general'" class="space-y-6">
              <div>
                <h2 class="text-lg font-semibold text-foreground">General</h2>
                <p class="text-sm text-muted-foreground">Project information and configuration.</p>
              </div>

              <!-- Project Info -->
              <section class="space-y-4">
                <h3 class="text-sm font-medium text-muted-foreground">Project Info</h3>
                <div class="space-y-3 rounded-lg border p-4">
                  <div class="flex items-center justify-between">
                    <span class="text-sm text-muted-foreground">Name</span>
                    <span class="text-sm font-medium text-foreground">{{ project.name }}</span>
                  </div>
                  <div v-if="project.createdAt" class="flex items-center justify-between">
                    <span class="text-sm text-muted-foreground">Created</span>
                    <span class="text-sm text-foreground">
                      {{ new Date(project.createdAt).toLocaleDateString() }}
                    </span>
                  </div>
                </div>
              </section>

              <Separator />

              <!-- Danger Zone -->
              <section class="space-y-4">
                <h3 class="text-sm font-medium text-destructive">Danger Zone</h3>
                <div class="rounded-lg border border-destructive/30 p-4">
                  <div class="flex items-center justify-between">
                    <div>
                      <p class="text-sm font-medium text-foreground">Delete Project</p>
                      <p class="text-xs text-muted-foreground">
                        Permanently delete this project and all its data.
</p>
                    </div>
                    <AlertDialog>
                      <AlertDialogTrigger as-child>
                        <Button variant="destructive" size="sm">
                          <Trash2 :size="14" class="mr-1" />
                          Delete
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Delete project</AlertDialogTitle>
                          <AlertDialogDescription>
                            This will permanently delete <strong>{{ project.name }}</strong>.
                            All environments, services, and deployments will be permanently deleted.
                            This action cannot be undone.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction
                            :disabled="deleting"
                            @click="handleDeleteProject"
                          >
                            {{ deleting ? 'Deleting...' : 'Delete' }}
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </div>
                </div>
              </section>
            </div>

            <!-- Environments -->
            <div v-if="activeSection === 'environments'" class="space-y-6">
              <div>
                <h2 class="text-lg font-semibold text-foreground">Environments</h2>
                <p class="text-sm text-muted-foreground">Manage environments for this project.</p>
              </div>

              <div v-if="project.environments?.length" class="overflow-hidden rounded-lg border">
                <div class="divide-y">
                  <div
                    v-for="env in project.environments"
                    :key="env.id"
                  >
                    <!-- Environment row -->
                    <div class="flex items-center justify-between px-4 py-3">
                      <button
                        class="flex items-center gap-2 text-left"
                        @click="toggleEnvExpand(env.name)"
                      >
                        <component
                          :is="expandedEnv === env.name ? ChevronDown : ChevronRight"
                          :size="14"
                          class="text-muted-foreground"
                        />
                        <span class="text-sm font-medium text-foreground">{{ env.name }}</span>
                        <span
                          v-if="env.ephemeral"
                          class="rounded-full bg-muted px-2 py-0.5 text-[10px] font-medium text-muted-foreground"
                        >
                          ephemeral
                        </span>
                        <span class="text-xs text-muted-foreground">
                          {{ env.resourceTier === 'PRODUCTION' ? 'Production' : 'Eco' }}
                        </span>
                      </button>
                      <Button
                        variant="ghost"
                        size="icon"
                        class="h-8 w-8 text-muted-foreground hover:text-destructive"
                        @click.stop="envToDelete = env.name"
                      >
                        <Trash2 :size="14" />
                      </Button>
                    </div>

                    <!-- Expanded resource panel -->
                    <div
                      v-if="expandedEnv === env.name"
                      class="border-t bg-muted/30 px-4 py-4"
                    >
                      <template v-if="envResources[env.name]?.loading">
                        <div class="space-y-3">
                          <Skeleton class="h-8 w-full" />
                          <Skeleton class="h-8 w-full" />
                          <Skeleton class="h-8 w-full" />
                        </div>
                      </template>
                      <template v-else-if="envResources[env.name]?.loaded">
                        <div class="space-y-5">
                          <!-- Tier -->
                          <div class="space-y-2">
                            <Label>Resource tier</Label>
                            <RadioGroup
                              :model-value="envResources[env.name]!.tier"
                              class="grid grid-cols-2 gap-3"
                              @update:model-value="envResources[env.name]!.tier = $event"
                            >
                              <label
                                class="flex cursor-pointer flex-col gap-1 rounded-lg border p-3 transition-colors"
                                :class="envResources[env.name]!.tier === 'ECO' ? 'border-primary bg-primary/5' : 'border-border'"
                              >
                                <div class="flex items-center gap-2">
                                  <RadioGroupItem value="ECO" />
                                  <span class="text-sm font-medium">Eco</span>
                                </div>
                                <p class="text-xs text-muted-foreground">
                                  Pay for what you use. Best for development, staging, and side projects.
                                </p>
                              </label>
                              <label
                                class="flex cursor-pointer flex-col gap-1 rounded-lg border p-3 transition-colors"
                                :class="envResources[env.name]!.tier === 'PRODUCTION' ? 'border-primary bg-primary/5' : 'border-border'"
                              >
                                <div class="flex items-center gap-2">
                                  <RadioGroupItem value="PRODUCTION" />
                                  <span class="text-sm font-medium">Production</span>
                                </div>
                                <p class="text-xs text-muted-foreground">
                                  Reserved resources. Best for production workloads with predictable load.
                                </p>
                              </label>
                            </RadioGroup>
                          </div>

                          <!-- ECO: no quota controls -->
                          <p
                            v-if="envResources[env.name]!.tier === 'ECO'"
                            class="text-sm text-muted-foreground"
                          >
                            Pay for what you use. No resource limits applied.
                          </p>

                          <!-- PRODUCTION: resource allocation controls -->
                          <template v-else>
                            <!-- CPU -->
                            <div class="space-y-2">
                              <div class="flex items-center justify-between">
                                <Label>CPU</Label>
                                <span class="text-sm font-medium">{{ formatCpu(envResources[env.name]!.cpuMillicores) }}</span>
                              </div>
                              <Slider
                                :model-value="[toIndex(cpuSteps, envResources[env.name]!.cpuMillicores)]"
                                :min="0"
                                :max="cpuSteps.length - 1"
                                :step="1"
                                @update:model-value="envResources[env.name]!.cpuMillicores = cpuSteps[$event?.[0] ?? 0]!"
                              />
                              <div class="flex justify-between text-[10px] text-muted-foreground">
                                <span v-for="s in cpuSteps" :key="s">{{ formatCpu(s) }}</span>
                              </div>
                            </div>

                            <!-- Memory -->
                            <div class="space-y-2">
                              <div class="flex items-center justify-between">
                                <Label>Memory</Label>
                                <span class="text-sm font-medium">{{ formatMB(envResources[env.name]!.memoryMB) }}</span>
                              </div>
                              <Slider
                                :model-value="[toIndex(memorySteps, envResources[env.name]!.memoryMB)]"
                                :min="0"
                                :max="memorySteps.length - 1"
                                :step="1"
                                @update:model-value="envResources[env.name]!.memoryMB = memorySteps[$event?.[0] ?? 0]!"
                              />
                              <div class="flex justify-between text-[10px] text-muted-foreground">
                                <span v-for="s in memorySteps" :key="s">{{ formatMB(s) }}</span>
                              </div>
                            </div>

                            <!-- Disk -->
                            <div class="space-y-2">
                              <div class="flex items-center justify-between">
                                <Label>Disk</Label>
                                <span class="text-sm font-medium">{{ formatMB(envResources[env.name]!.diskMB) }}</span>
                              </div>
                              <Slider
                                :model-value="[toIndex(diskSteps, envResources[env.name]!.diskMB)]"
                                :min="0"
                                :max="diskSteps.length - 1"
                                :step="1"
                                @update:model-value="envResources[env.name]!.diskMB = diskSteps[$event?.[0] ?? 0]!"
                              />
                              <div class="flex justify-between text-[10px] text-muted-foreground">
                                <span v-for="s in diskSteps" :key="s">{{ formatMB(s) }}</span>
                              </div>
                            </div>
                          </template>

                          <!-- Save -->
                          <div class="flex justify-end">
                            <Button
                              size="sm"
                              :disabled="envResources[env.name]!.saving"
                              @click="handleSaveResources(env.name)"
                            >
                              {{ envResources[env.name]!.saving ? 'Saving...' : 'Save resources' }}
                            </Button>
                          </div>
                        </div>
                      </template>
                    </div>
                  </div>
                </div>
              </div>

              <p
                v-else
                class="text-sm text-muted-foreground"
              >
                No environments found.
              </p>

              <!-- Delete confirmation dialog -->
              <AlertDialog :open="!!envToDelete">
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Delete environment "{{ envToDelete }}"?</AlertDialogTitle>
                    <AlertDialogDescription>
                      This will permanently delete the environment and all its deployments.
                      This action cannot be undone.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel @click="envToDelete = null">Cancel</AlertDialogCancel>
                    <Button
                      variant="destructive"
                      :disabled="deletingEnv"
                      @click="handleDeleteEnvironment"
                    >
                      {{ deletingEnv ? 'Deleting...' : 'Delete' }}
                    </Button>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </div>

            <!-- Variables -->
            <div v-if="activeSection === 'variables'" class="space-y-6">
              <div>
                <h2 class="text-lg font-semibold text-foreground">Variables</h2>
                <p class="text-sm text-muted-foreground">
                  Shared variables that services can reference in their configuration.
                </p>
              </div>

              <SharedVariablesEditor :project-id="projectId" />
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
