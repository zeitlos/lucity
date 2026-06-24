<script setup lang="ts">
import { useQuery } from '@vue/apollo-composable';
import { useRoute, useRouter } from 'vue-router';
import { computed, ref, watch } from 'vue';
import { Plus, Box, FolderGit2 } from '@lucide/vue';
import GithubIcon from '@/components/GithubIcon.vue';
import { graphql } from '@/gql';

const ProjectsDocument = graphql(`
  query Projects {
    projects {
      id
      name
      environments {
        id
        name
        resourceTier
        services {
          id
          name
          sourceUrl
        }
      }
    }
  }
`);

const GitHubConnectedDocument = graphql(`
  query GitHubConnected {
    githubConnected
  }
`);
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import EmptyState from '@/components/EmptyState.vue';
import CreateCommandPalette from '@/components/CreateCommandPalette.vue';
import WelcomeCard from '@/components/WelcomeCard.vue';
import { useOnboarding } from '@/composables/useOnboarding';
import { useGitHubInstall } from '@/composables/useGitHubInstall';

const route = useRoute();
const router = useRouter();

const { result, loading, error } = useQuery(ProjectsDocument);
const { result: ghResult } = useQuery(GitHubConnectedDocument);

const projects = computed(() => result.value?.projects ?? []);
const githubConnected = computed(() => ghResult.value?.githubConnected ?? false);
const paletteOpen = ref(false);
const initialPaletteView = ref<'main' | 'github-repos'>('main');

const { isWelcome, dismissWelcome } = useOnboarding();
const { openInstallPopup } = useGitHubInstall();

function handleWelcomeDismiss() {
  dismissWelcome();
}

function handleCreateProject() {
  dismissWelcome();
  paletteOpen.value = true;
}

function handleImportGitHub() {
  dismissWelcome();
  initialPaletteView.value = 'github-repos';
  paletteOpen.value = true;
}

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

function allServices(project: { environments: { services?: { id: string; name: string; sourceUrl: string }[] }[] }) {
  const seen = new Set<string>();
  const out: { name: string; sourceUrl: string }[] = [];
  for (const env of project.environments) {
    for (const svc of env.services ?? []) {
      if (!seen.has(svc.name)) {
        seen.add(svc.name);
        out.push(svc);
      }
    }
  }
  return out;
}

function uniqueRepoCount(services: { sourceUrl: string }[]): number {
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

    <WelcomeCard
      v-else-if="isWelcome && projects.length === 0"
      @dismiss="handleWelcomeDismiss"
      @create-project="handleCreateProject"
      @import-github="handleImportGitHub"
    />

    <template v-else-if="projects.length === 0">
      <EmptyState
        v-if="!githubConnected"
        title="Connect GitHub"
        description="Connect your GitHub account to import repositories and deploy your first project."
        pattern="dots"
      >
        <template #action>
          <Button @click="openInstallPopup()">
            <GithubIcon :size="16" class="mr-2" primary />
            Connect GitHub
          </Button>
        </template>
      </EmptyState>

      <EmptyState
        v-else
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
    </template>

    <template v-else>
      <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      <RouterLink
        v-for="project in projects"
        :key="project.id"
        class="block text-left"
        :to="{ name: 'project', params: { projectId: project.id }}"
      >
        <Card class="transition-shadow hover:shadow-md">
          <CardHeader>
            <CardTitle class="text-lg">{{ project.name }}</CardTitle>
            <CardDescription class="flex items-center gap-3">
              <span class="flex items-center gap-1">
                <Box :size="12" />
                {{ allServices(project).length }} service{{ allServices(project).length !== 1 ? 's' : '' }}
              </span>
              <span class="flex items-center gap-1">
                <FolderGit2 :size="12" />
                {{ uniqueRepoCount(allServices(project)) }} repo{{ uniqueRepoCount(allServices(project)) !== 1 ? 's' : '' }}
              </span>
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex items-center gap-2 text-xs text-muted-foreground">
              {{ project.environments.length }} environment{{ project.environments.length !== 1 ? 's' : '' }}
            </div>
          </CardContent>
        </Card>
      </RouterLink>
      </div>
    </template>

    <CreateCommandPalette
      v-model:open="paletteOpen"
      context="projects"
      :initial-view="initialPaletteView"
    />
  </div>
</template>
