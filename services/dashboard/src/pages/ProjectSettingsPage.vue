<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { ArrowLeft, Trash2 } from 'lucide-vue-next';
import { ProjectQuery, DeleteProjectMutation } from '@/graphql/projects';
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
const { setEnvironments } = useEnvironment();

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
