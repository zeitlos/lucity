<script setup lang="ts">
import { computed, ref, reactive, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { ArrowLeft, Trash2, ChevronDown, ChevronRight } from 'lucide-vue-next';
import { graphql } from '@/gql';
import { ResourceTier } from '@/gql/graphql';

const ProjectDocument = graphql(`
  query ProjectSettings($id: ProjectID!) {
    project(id: $id) {
      id
      name
      environments {
        id
        name
        resourceTier
      }
    }
  }
`);

const DeleteProjectDocument = graphql(`
  mutation DeleteProject($id: ProjectID!) {
    deleteProject(id: $id)
  }
`);

const DeleteEnvironmentDocument = graphql(`
  mutation DeleteEnvironment($environment: EnvironmentID!) {
    deleteEnvironment(environment: $environment)
  }
`);

const EnvironmentResourcesDocument = graphql(`
  query EnvironmentResources($environment: EnvironmentID!) {
    environmentResources(environment: $environment) {
      tier
      allocation {
        cpuMillicores
        memoryMB
        diskMB
      }
    }
  }
`);

const SetEnvironmentResourcesDocument = graphql(`
  mutation SetEnvironmentResources($input: SetEnvironmentResourcesInput!) {
    setEnvironmentResources(input: $input) {
      id
      resourceTier
    }
  }
`);
import { apolloClient } from '@/lib/apollo';
import { useEnvironment, type Environment } from '@/composables/useEnvironment';
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
const projectId = computed(() => {
  const id = route.params.projectId;
  return Array.isArray(id) ? id[0]! : (id as string);
});

const { result, loading } = useQuery(ProjectDocument, () => ({
  id: projectId.value,
}));

const project = computed(() => result.value?.project);

const { setEnvironments, environments, activeEnvironment, setEnvironment } = useEnvironment();

watch(
  () => project.value?.environments,
  (envs) => {
    if (envs) {
      const shells: Environment[] = envs.map(e => ({
        id: e.id,
        name: e.name,
        resourceTier: e.resourceTier,
        services: [],
        databases: [],
        keyValueStores: [],
      }));
      setEnvironments(shells);
    }
  },
  { immediate: true },
);

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
      params: { projectId: projectId.value, section: val === 'general' ? undefined : val },
      query: route.query,
    });
  },
});

// Delete project
const { mutate: deleteProjectMutate, loading: deleting } = useMutation(DeleteProjectDocument);

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
  return `${m / 1000} vCPU`;
}

function formatMB(mb: number) {
  return mb >= 1024 ? `${mb / 1024} GB` : `${mb} MB`;
}

// Environment resources — keyed by EnvironmentID
interface EnvResourceState {
  loading: boolean;
  loaded: boolean;
  saving: boolean;
  tier: ResourceTier;
  cpuMillicores: number;
  memoryMB: number;
  diskMB: number;
}

const expandedEnvId = ref<string | null>(null);
const envResources: Record<string, EnvResourceState> = reactive({});

const { mutate: setResourcesMutate } = useMutation(SetEnvironmentResourcesDocument);

async function toggleEnvExpand(envId: string) {
  if (expandedEnvId.value === envId) {
    expandedEnvId.value = null;
    return;
  }
  expandedEnvId.value = envId;

  if (envResources[envId]?.loaded) return;

  envResources[envId] = {
    loading: true,
    loaded: false,
    saving: false,
    tier: ResourceTier.Eco,
    cpuMillicores: 1000,
    memoryMB: 1024,
    diskMB: 1024,
  };

  try {
    const { data } = await apolloClient.query({
      query: EnvironmentResourcesDocument,
      variables: { environment: envId },
      fetchPolicy: 'network-only',
    });
    if (data?.environmentResources) {
      const r = data.environmentResources;
      envResources[envId]!.tier = r.tier;
      envResources[envId]!.cpuMillicores = r.allocation.cpuMillicores;
      envResources[envId]!.memoryMB = r.allocation.memoryMB;
      envResources[envId]!.diskMB = r.allocation.diskMB;
    }
  } catch {
    // No resources set yet — keep defaults
  } finally {
    envResources[envId]!.loading = false;
    envResources[envId]!.loaded = true;
  }
}

// Auto-expand environment from query param (?env=name)
watch(
  () => [route.query.env, project.value?.environments],
  () => {
    const envName = route.query.env as string | undefined;
    if (!envName || !project.value?.environments) return;
    const match = project.value.environments.find(e => e.name === envName);
    if (match && expandedEnvId.value !== match.id) {
      toggleEnvExpand(match.id);
    }
  },
  { immediate: true },
);

async function handleSaveResources(envId: string) {
  const state = envResources[envId];
  if (!state) return;

  state.saving = true;
  try {
    await setResourcesMutate({
      input: {
        environment: envId,
        tier: state.tier,
        cpuMillicores: state.cpuMillicores,
        memoryMB: state.memoryMB,
        diskMB: state.diskMB,
      },
    });
    toast.success('Resources updated');
  } catch (e: unknown) {
    errorToast('Failed to update resources', { description: errorMessage(e) });
  } finally {
    state.saving = false;
  }
}

// Delete environment
const { mutate: deleteEnvironmentMutate, loading: deletingEnv } = useMutation(DeleteEnvironmentDocument, {
  refetchQueries: () => [{ query: ProjectDocument, variables: { id: projectId.value } }],
});
const envToDelete = ref<{ id: string; name: string } | null>(null);

async function handleDeleteEnvironment() {
  const target = envToDelete.value;
  if (!target) return;

  try {
    const res = await deleteEnvironmentMutate({ environment: target.id });

    if (res?.errors?.length) {
      errorToast('Failed to delete environment', {
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
      });
      return;
    }

    if (activeEnvironment.value?.id === target.id) {
      const remaining = environments.value.filter(e => e.id !== target.id);
      if (remaining.length > 0) {
        setEnvironment(remaining[0]!);
      }
    }

    toast.success(`Environment "${target.name}" deleted`);
  } catch (e: unknown) {
    errorToast('Failed to delete environment', { description: errorMessage(e) });
  } finally {
    envToDelete.value = null;
  }
}
</script>

<template>
  <div class="flex h-[calc(100vh-52px-0.75rem)] flex-col">
    <div v-if="loading" class="flex flex-1 items-center justify-center">
      <div class="space-y-4 text-center">
        <Skeleton class="mx-auto h-8 w-48" />
        <Skeleton class="mx-auto h-4 w-64" />
      </div>
    </div>

    <template v-else-if="project">
      <div class="flex flex-1 overflow-hidden p-3">
        <div class="mx-auto flex w-full max-w-4xl gap-6 overflow-hidden rounded-lg border bg-card shadow-sm">
          <!-- Sidebar -->
          <nav class="w-48 shrink-0 border-r p-4">
            <div class="mb-4">
              <RouterLink
                class="flex items-center gap-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
                :to="{ name: 'project', params: { projectId } }"
              >
                <ArrowLeft :size="12" />
                Back to project
              </RouterLink>
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

              <section class="space-y-4">
                <h3 class="text-sm font-medium text-muted-foreground">Project Info</h3>
                <div class="space-y-3 rounded-lg border p-4">
                  <div class="flex items-center justify-between">
                    <span class="text-sm text-muted-foreground">Name</span>
                    <span class="text-sm font-medium text-foreground">{{ project.name }}</span>
                  </div>
                  <div class="flex items-center justify-between">
                    <span class="text-sm text-muted-foreground">ID</span>
                    <span class="font-mono text-sm text-foreground">{{ project.id }}</span>
                  </div>
                </div>
              </section>

              <Separator />

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
                    <div class="flex items-center justify-between px-4 py-3">
                      <button
                        class="flex items-center gap-2 text-left"
                        @click="toggleEnvExpand(env.id)"
                      >
                        <component
                          :is="expandedEnvId === env.id ? ChevronDown : ChevronRight"
                          :size="14"
                          class="text-muted-foreground"
                        />
                        <span class="text-sm font-medium text-foreground">{{ env.name }}</span>
                        <span class="text-xs text-muted-foreground">
                          {{ env.resourceTier === ResourceTier.Production ? 'Production' : 'Eco' }}
                        </span>
                      </button>
                      <Button
                        variant="ghost"
                        size="icon"
                        class="h-8 w-8 text-muted-foreground hover:text-destructive"
                        @click.stop="envToDelete = { id: env.id, name: env.name }"
                      >
                        <Trash2 :size="14" />
                      </Button>
                    </div>

                    <div
                      v-if="expandedEnvId === env.id"
                      class="border-t bg-muted/30 px-4 py-4"
                    >
                      <template v-if="envResources[env.id]?.loading">
                        <div class="space-y-3">
                          <Skeleton class="h-8 w-full" />
                          <Skeleton class="h-8 w-full" />
                          <Skeleton class="h-8 w-full" />
                        </div>
                      </template>
                      <template v-else-if="envResources[env.id]?.loaded">
                        <div class="space-y-5">
                          <div class="space-y-2">
                            <Label>Resource tier</Label>
                            <RadioGroup
                              :model-value="envResources[env.id]!.tier"
                              class="grid grid-cols-2 gap-3"
                              @update:model-value="envResources[env.id]!.tier = $event as ResourceTier"
                            >
                              <label
                                class="flex cursor-pointer flex-col gap-1 rounded-lg border p-3 transition-colors"
                                :class="envResources[env.id]!.tier === ResourceTier.Eco ? 'border-primary bg-primary/5' : 'border-border'"
                              >
                                <div class="flex items-center gap-2">
                                  <RadioGroupItem :value="ResourceTier.Eco" />
                                  <span class="text-sm font-medium">Eco</span>
                                </div>
                                <p class="text-xs text-muted-foreground">
                                  Pay for what you use. Best for development, staging, and side projects.
                                </p>
                              </label>
                              <label
                                class="flex cursor-pointer flex-col gap-1 rounded-lg border p-3 transition-colors"
                                :class="envResources[env.id]!.tier === ResourceTier.Production ? 'border-primary bg-primary/5' : 'border-border'"
                              >
                                <div class="flex items-center gap-2">
                                  <RadioGroupItem :value="ResourceTier.Production" />
                                  <span class="text-sm font-medium">Production</span>
                                </div>
                                <p class="text-xs text-muted-foreground">
                                  Reserved resources. Best for production workloads with predictable load.
                                </p>
                              </label>
                            </RadioGroup>
                          </div>

                          <p
                            v-if="envResources[env.id]!.tier === ResourceTier.Eco"
                            class="text-sm text-muted-foreground"
                          >
                            Cheaper. Best for development, staging, and side projects. Performance can vary.
                          </p>

                          <template v-else>
                            <div class="space-y-2">
                              <div class="flex items-center justify-between">
                                <Label>CPU</Label>
                                <span class="text-sm font-medium">{{ formatCpu(envResources[env.id]!.cpuMillicores) }}</span>
                              </div>
                              <Slider
                                :model-value="[toIndex(cpuSteps, envResources[env.id]!.cpuMillicores)]"
                                :min="0"
                                :max="cpuSteps.length - 1"
                                :step="1"
                                @update:model-value="envResources[env.id]!.cpuMillicores = cpuSteps[$event?.[0] ?? 0]!"
                              />
                              <div class="flex justify-between text-[10px] text-muted-foreground">
                                <span v-for="s in cpuSteps" :key="s">{{ formatCpu(s) }}</span>
                              </div>
                            </div>

                            <div class="space-y-2">
                              <div class="flex items-center justify-between">
                                <Label>Memory</Label>
                                <span class="text-sm font-medium">{{ formatMB(envResources[env.id]!.memoryMB) }}</span>
                              </div>
                              <Slider
                                :model-value="[toIndex(memorySteps, envResources[env.id]!.memoryMB)]"
                                :min="0"
                                :max="memorySteps.length - 1"
                                :step="1"
                                @update:model-value="envResources[env.id]!.memoryMB = memorySteps[$event?.[0] ?? 0]!"
                              />
                              <div class="flex justify-between text-[10px] text-muted-foreground">
                                <span v-for="s in memorySteps" :key="s">{{ formatMB(s) }}</span>
                              </div>
                            </div>

                            <div class="space-y-2">
                              <div class="flex items-center justify-between">
                                <Label>Disk</Label>
                                <span class="text-sm font-medium">{{ formatMB(envResources[env.id]!.diskMB) }}</span>
                              </div>
                              <Slider
                                :model-value="[toIndex(diskSteps, envResources[env.id]!.diskMB)]"
                                :min="0"
                                :max="diskSteps.length - 1"
                                :step="1"
                                @update:model-value="envResources[env.id]!.diskMB = diskSteps[$event?.[0] ?? 0]!"
                              />
                              <div class="flex justify-between text-[10px] text-muted-foreground">
                                <span v-for="s in diskSteps" :key="s">{{ formatMB(s) }}</span>
                              </div>
                            </div>
                          </template>

                          <div class="flex justify-end">
                            <Button
                              size="sm"
                              :disabled="envResources[env.id]!.saving"
                              @click="handleSaveResources(env.id)"
                            >
                              {{ envResources[env.id]!.saving ? 'Saving...' : 'Save resources' }}
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

              <AlertDialog :open="!!envToDelete">
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Delete environment "{{ envToDelete?.name }}"?</AlertDialogTitle>
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

              <SharedVariablesEditor />
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
