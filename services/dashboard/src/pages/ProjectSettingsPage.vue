<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { ArrowLeft, Trash2 } from 'lucide-vue-next';
import { ProjectQuery, DeleteProjectMutation, DeleteEnvironmentMutation } from '@/graphql/projects';
import { apolloClient } from '@/lib/apollo';
import { useEnvironment } from '@/composables/useEnvironment';
import SharedVariablesEditor from '@/components/SharedVariablesEditor.vue';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
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

// Delete environment
const { mutate: deleteEnvironmentMutate, loading: deletingEnv } = useMutation(DeleteEnvironmentMutation, {
  refetchQueries: () => [{ query: ProjectQuery, variables: { id: projectId.value } }],
});
const envToDelete = ref<string | null>(null);

async function handleDeleteEnvironment() {
  if (!envToDelete.value) return;

  try {
    const res = await deleteEnvironmentMutate({
      projectId: projectId.value,
      environment: envToDelete.value,
    });

    if (res?.errors?.length) {
      toast.error('Failed to delete environment', {
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
      });
      return;
    }

    // If the deleted environment was active, switch to another one
    if (activeEnvironment.value?.name === envToDelete.value) {
      const remaining = environments.value.filter(e => e.name !== envToDelete.value);
      if (remaining.length > 0) {
        setEnvironment(remaining[0]);
      }
    }

    toast.success(`Environment "${envToDelete.value}" deleted`);
    envToDelete.value = null;
  } catch (e: unknown) {
    toast.error('Failed to delete environment', { description: errorMessage(e) });
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
                        Permanently delete this project and its GitOps repository.
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
                            This will permanently delete <strong>{{ project.name }}</strong> and its
                            GitOps repository. All environments and deployments will be removed.
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

              <div class="overflow-hidden rounded-lg border">
                <div class="divide-y">
                  <div
                    v-for="env in project.environments"
                    :key="env.id"
                    class="flex items-center justify-between px-4 py-3"
                  >
                    <div class="flex items-center gap-3">
                      <span class="text-sm font-medium text-foreground">{{ env.name }}</span>
                      <span
                        v-if="env.ephemeral"
                        class="rounded-full bg-muted px-2 py-0.5 text-[10px] font-medium text-muted-foreground"
                      >
                        ephemeral
                      </span>
                    </div>
                    <AlertDialog
                      :open="envToDelete === env.name"
                      @update:open="(open: boolean) => { if (!open) envToDelete = null; }"
                    >
                      <AlertDialogTrigger as-child>
                        <Button
                          variant="ghost"
                          size="icon"
                          class="h-8 w-8 text-muted-foreground hover:text-destructive"
                          @click="envToDelete = env.name"
                        >
                          <Trash2 :size="14" />
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Delete environment "{{ env.name }}"?</AlertDialogTitle>
                          <AlertDialogDescription>
                            This will remove the environment and its ArgoCD application.
                            All deployments in this environment will be deleted.
                            This action cannot be undone.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction
                            :disabled="deletingEnv"
                            class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            @click="handleDeleteEnvironment"
                          >
                            {{ deletingEnv ? 'Deleting...' : 'Delete' }}
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </div>
                </div>
              </div>

              <p
                v-if="!project.environments?.length"
                class="text-sm text-muted-foreground"
              >
                No environments found.
              </p>
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
