<script setup lang="ts">
import { useQuery } from '@vue/apollo-composable';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { computed, ref, watch } from 'vue';
import { Plus, Github, Box } from 'lucide-vue-next';
import { ProjectsQuery } from '@/graphql/projects';
import { GitHubConnectedQuery } from '@/graphql/github';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import EmptyState from '@/components/EmptyState.vue';
import CreateCommandPalette from '@/components/CreateCommandPalette.vue';

const route = useRoute();
const router = useRouter();

const { result, loading, error } = useQuery(ProjectsQuery);
const { result: ghResult } = useQuery(GitHubConnectedQuery);

const projects = computed(() => result.value?.projects ?? []);
const githubConnected = computed(() => ghResult.value?.githubConnected ?? false);
const paletteOpen = ref(false);
const initialPaletteView = ref<'main' | 'github-repos'>('main');

// Auto-open palette on github-repos view when returning from GitHub account connection
watch(() => route.query.github, (val) => {
  if (val === 'account_connected') {
    initialPaletteView.value = 'github-repos';
    paletteOpen.value = true;
    router.replace({ query: {} });
  }
}, { immediate: true });

watch(paletteOpen, (open) => {
  if (!open) initialPaletteView.value = 'main';
});

function envStatusColor(environments: { syncStatus: string }[]) {
  if (environments.length === 0) return 'bg-muted-foreground/50';
  const hasDegraded = environments.some(e => e.syncStatus === 'DEGRADED');
  if (hasDegraded) return 'bg-red-500';
  const allSynced = environments.every(e => e.syncStatus === 'SYNCED');
  if (allSynced) return 'bg-green-500';
  return 'bg-yellow-500';
}

function uniqueRepoCount(services: { sourceUrl?: string }[]): number {
  const urls = services.filter(s => s.sourceUrl).map(s => s.sourceUrl);
  return new Set(urls).size;
}
</script>

<template>
  <div class="p-8">
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="font-serif text-3xl text-foreground">Projects</h1>
        <p class="mt-1 text-sm text-muted-foreground">Your deployed applications.</p>
      </div>
      <Button @click="paletteOpen = true">
        <Plus :size="16" class="mr-2" />
        New
      </Button>
    </div>

    <div v-if="loading" class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      <Card v-for="i in 3" :key="i">
        <CardHeader>
          <Skeleton class="h-5 w-32" />
          <Skeleton class="mt-2 h-4 w-48" />
        </CardHeader>
        <CardContent>
          <Skeleton class="h-4 w-full" />
          <Skeleton class="mt-2 h-4 w-24" />
        </CardContent>
      </Card>
    </div>

    <div
      v-else-if="error"
      class="rounded-lg border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive"
    >
      Failed to load projects: {{ error.message }}
    </div>

    <EmptyState
      v-else-if="projects.length === 0 && !githubConnected"
      title="Connect GitHub"
      description="Connect your GitHub account to import repositories and deploy your first project."
      pattern="dots"
    >
      <template #action>
        <a href="/auth/github/connect" class="inline-flex">
          <Button>
            <Github :size="16" class="mr-2" />
            Connect GitHub
          </Button>
        </a>
      </template>
    </EmptyState>

    <EmptyState
      v-else-if="projects.length === 0"
      title="No projects yet"
      description="Get started by connecting a GitHub repository."
      pattern="dots"
    >
      <template #action>
        <Button @click="paletteOpen = true">
          <Plus :size="16" class="mr-2" />
          New Project
        </Button>
      </template>
    </EmptyState>

    <div v-else class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      <RouterLink
        v-for="project in projects"
        :key="project.id"
        :to="{ name: 'project', params: { id: project.id } }"
        class="block"
      >
        <Card class="transition-shadow hover:shadow-md">
          <CardHeader>
            <CardTitle class="text-lg">{{ project.name }}</CardTitle>
            <CardDescription class="flex items-center gap-3">
              <span v-if="project.services?.length" class="flex items-center gap-1">
                <Box :size="12" />
                {{ project.services.length }} service{{ project.services.length !== 1 ? 's' : '' }}
              </span>
              <span v-if="uniqueRepoCount(project.services ?? [])" class="flex items-center gap-1">
                <Github :size="12" />
                {{ uniqueRepoCount(project.services) }} repo{{ uniqueRepoCount(project.services) !== 1 ? 's' : '' }}
              </span>
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex items-center gap-2 text-xs text-muted-foreground">
              <span
                :class="['h-2 w-2 rounded-full', envStatusColor(project.environments)]"
              />
              {{ project.environments.length }} environment{{ project.environments.length !== 1 ? 's' : '' }}
            </div>
          </CardContent>
        </Card>
      </RouterLink>
    </div>

    <CreateCommandPalette
      v-model:open="paletteOpen"
      context="projects"
      :initial-view="initialPaletteView"
    />
  </div>
</template>
