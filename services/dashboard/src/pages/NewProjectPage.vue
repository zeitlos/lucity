<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { ArrowLeft, Lock, Globe, Search, FolderGit2 } from 'lucide-vue-next';
import { GitHubRepositoriesQuery } from '@/graphql/github';
import { CreateProjectMutation } from '@/graphql/projects';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from '@/components/ui/sonner';
import EmptyState from '@/components/EmptyState.vue';

const router = useRouter();

const { result, loading, error } = useQuery(GitHubRepositoriesQuery);
const { mutate: createProject, loading: creating, onError: onCreateError } = useMutation(CreateProjectMutation);
onCreateError((error) => {
  toast.error('Failed to create project', { description: error.message });
});

const search = ref('');
const selectedRepo = ref<{ fullName: string; htmlUrl: string } | null>(null);

const filteredRepos = computed(() => {
  const repos = result.value?.githubRepositories ?? [];
  if (!search.value) return repos;
  const q = search.value.toLowerCase();
  return repos.filter((r: { fullName: string }) => r.fullName.toLowerCase().includes(q));
});

function selectRepo(repo: { fullName: string; htmlUrl: string }) {
  selectedRepo.value = repo;
}

async function handleCreate() {
  if (!selectedRepo.value) return;

  const res = await createProject({
    input: {
      name: selectedRepo.value.fullName,
      sourceUrl: selectedRepo.value.htmlUrl,
    },
  });

  if (res?.data?.createProject) {
    router.push({ name: 'project', params: { id: res.data.createProject.id } });
  }
}
</script>

<template>
  <div class="p-8">
    <div class="mb-8">
      <Button
        variant="ghost"
        size="sm"
        class="mb-4"
        @click="router.push({ name: 'projects' })"
      >
        <ArrowLeft :size="16" class="mr-2" />
        Back to Projects
      </Button>
      <h1 class="text-2xl font-semibold text-foreground">New Project</h1>
      <p class="mt-1 text-sm text-muted-foreground">
        Select a repository to create a project from.
      </p>
    </div>

    <!-- Selected repo confirmation -->
    <div v-if="selectedRepo" class="mb-8">
      <Card class="border-primary/30 bg-primary/10">
        <CardContent class="flex items-center justify-between pt-6">
          <div>
            <p class="text-sm text-muted-foreground">Creating project from</p>
            <p class="font-medium text-foreground">{{ selectedRepo.fullName }}</p>
          </div>
          <div class="flex gap-2">
            <Button variant="outline" size="sm" @click="selectedRepo = null">
              Change
            </Button>
            <Button
              size="sm"
              :disabled="creating"
              @click="handleCreate"
            >
              {{ creating ? 'Creating...' : 'Create Project' }}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Search -->
    <div v-if="!selectedRepo" class="relative mb-6">
      <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
      <Input
        v-model="search"
        placeholder="Search repositories..."
        class="pl-10"
      />
    </div>

    <!-- Loading -->
    <div v-if="loading && !selectedRepo" class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <Card v-for="i in 6" :key="i">
        <CardHeader>
          <Skeleton class="h-5 w-40" />
        </CardHeader>
        <CardContent>
          <Skeleton class="h-4 w-24" />
        </CardContent>
      </Card>
    </div>

    <!-- Error -->
    <div
      v-else-if="error && !selectedRepo"
      class="rounded-lg border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive"
    >
      Failed to load repositories: {{ error.message }}
    </div>

    <!-- Repo list -->
    <div
      v-else-if="!selectedRepo"
      class="grid gap-4 md:grid-cols-2 lg:grid-cols-3"
    >
      <Card
        v-for="repo in filteredRepos"
        :key="repo.id"
        class="cursor-pointer transition-shadow hover:shadow-md"
        @click="selectRepo(repo)"
      >
        <CardHeader class="pb-2">
          <CardTitle class="flex items-center gap-2 text-base">
            <component :is="repo.private ? Lock : Globe" :size="14" class="text-muted-foreground" />
            {{ repo.name }}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p class="text-sm text-muted-foreground">{{ repo.fullName }}</p>
          <div class="mt-2 flex items-center gap-2">
            <Badge variant="outline">{{ repo.defaultBranch }}</Badge>
            <Badge v-if="repo.private" variant="secondary">Private</Badge>
          </div>
        </CardContent>
      </Card>

      <div v-if="filteredRepos.length === 0" class="col-span-full">
        <EmptyState
          v-if="search"
          :icon="Search"
          title="No results"
          :description="`No repositories matching &quot;${search}&quot;.`"
        />
        <EmptyState
          v-else
          :icon="FolderGit2"
          title="No repositories found"
          description="Make sure the Lucity GitHub App is installed and has access to your repositories."
        />
      </div>
    </div>
  </div>
</template>
