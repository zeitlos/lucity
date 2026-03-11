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
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { Slider } from '@/components/ui/slider';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
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
import { toast } from '@/components/ui/sonner';
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

// Settings sections
const activeSection = ref('general');

const sections = [
  { id: 'general', label: 'General' },
  { id: 'environments', label: 'Environments' },
  { id: 'variables', label: 'Variables' },
];

// Delete project
const { mutate: deleteProjectMutate, loading: deleting } = useMutation(DeleteProjectMutation);

async function handleDeleteProject() {
  try {
    const res = await deleteProjectMutate({ id: projectId.value });

    if (res?.errors?.length) {
      toast.error('Failed to delete project', {
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
      });
      return;
    }

    apolloClient.cache.evict({ id: `Project:${projectId.value}` });
    apolloClient.cache.gc();

    toast.success('Project deleted');
    router.push({ name: 'projects' });
  } catch (e: unknown) {
    toast.error('Failed to delete project', { description: errorMessage(e) });
  }
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
    toast.error('Failed to update resources', { description: errorMessage(e) });
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
      toast.error('Failed to delete environment', {
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
    toast.error('Failed to delete environment', { description: errorMessage(e) });
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
                        <Badge
                          v-if="envResources[env.name]?.loaded"
                          variant="secondary"
                          class="text-[10px]"
                        >
                          {{ envResources[env.name]!.tier === 'PRODUCTION' ? 'Production' : 'Eco' }}
                        </Badge>
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
                            <Select v-model="envResources[env.name]!.tier">
                              <SelectTrigger class="w-48">
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent>
                                <SelectItem value="ECO">
                                  <div>
                                    <span class="font-medium">Eco</span>
                                    <span class="ml-1.5 text-xs text-muted-foreground">Shared, billed by usage</span>
                                  </div>
                                </SelectItem>
                                <SelectItem value="PRODUCTION">
                                  <div>
                                    <span class="font-medium">Production</span>
                                    <span class="ml-1.5 text-xs text-muted-foreground">Dedicated, billed by allocation</span>
                                  </div>
                                </SelectItem>
                              </SelectContent>
                            </Select>
                          </div>

                          <!-- CPU -->
                          <div class="space-y-2">
                            <div class="flex items-center justify-between">
                              <Label>CPU</Label>
                              <div class="flex items-center gap-1.5">
                                <Input
                                  type="number"
                                  :model-value="envResources[env.name]!.cpuMillicores"
                                  class="h-7 w-20 text-right text-xs"
                                  :min="100"
                                  :max="32000"
                                  @update:model-value="envResources[env.name]!.cpuMillicores = Number($event)"
                                />
                                <span class="text-xs text-muted-foreground">m</span>
                              </div>
                            </div>
                            <Slider
                              :model-value="[envResources[env.name]!.cpuMillicores]"
                              :min="100"
                              :max="32000"
                              :step="100"
                              @update:model-value="envResources[env.name]!.cpuMillicores = $event?.[0] ?? envResources[env.name]!.cpuMillicores"
                            />
                            <p class="text-[11px] text-muted-foreground">
                              {{ (envResources[env.name]!.cpuMillicores / 1000).toFixed(1) }} vCPU
                            </p>
                          </div>

                          <!-- Memory -->
                          <div class="space-y-2">
                            <div class="flex items-center justify-between">
                              <Label>Memory</Label>
                              <div class="flex items-center gap-1.5">
                                <Input
                                  type="number"
                                  :model-value="envResources[env.name]!.memoryMB"
                                  class="h-7 w-20 text-right text-xs"
                                  :min="128"
                                  :max="65536"
                                  @update:model-value="envResources[env.name]!.memoryMB = Number($event)"
                                />
                                <span class="text-xs text-muted-foreground">MB</span>
                              </div>
                            </div>
                            <Slider
                              :model-value="[envResources[env.name]!.memoryMB]"
                              :min="128"
                              :max="65536"
                              :step="128"
                              @update:model-value="envResources[env.name]!.memoryMB = $event?.[0] ?? envResources[env.name]!.memoryMB"
                            />
                            <p class="text-[11px] text-muted-foreground">
                              {{ (envResources[env.name]!.memoryMB / 1024).toFixed(1) }} GB
                            </p>
                          </div>

                          <!-- Disk -->
                          <div class="space-y-2">
                            <div class="flex items-center justify-between">
                              <Label>Disk</Label>
                              <div class="flex items-center gap-1.5">
                                <Input
                                  type="number"
                                  :model-value="envResources[env.name]!.diskMB"
                                  class="h-7 w-20 text-right text-xs"
                                  :min="0"
                                  :max="102400"
                                  @update:model-value="envResources[env.name]!.diskMB = Number($event)"
                                />
                                <span class="text-xs text-muted-foreground">MB</span>
                              </div>
                            </div>
                            <Slider
                              :model-value="[envResources[env.name]!.diskMB]"
                              :min="0"
                              :max="102400"
                              :step="512"
                              @update:model-value="envResources[env.name]!.diskMB = $event?.[0] ?? envResources[env.name]!.diskMB"
                            />
                            <p class="text-[11px] text-muted-foreground">
                              {{ (envResources[env.name]!.diskMB / 1024).toFixed(1) }} GB
                            </p>
                          </div>

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
