<script setup lang="ts">
import { useQuery } from '@vue/apollo-composable';
import { RouterLink } from 'vue-router';
import { Plus, GitBranch, ExternalLink } from 'lucide-vue-next';
import { ProjectsQuery } from '@/graphql/projects';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';

const { result, loading, error } = useQuery(ProjectsQuery);

function syncStatusVariant(status: string) {
  switch (status) {
    case 'SYNCED': return 'default';
    case 'PROGRESSING': return 'secondary';
    case 'OUT_OF_SYNC': return 'outline';
    case 'DEGRADED': return 'destructive';
    default: return 'outline';
  }
}
</script>

<template>
  <div class="p-8">
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">Projects</h1>
        <p class="mt-1 text-sm text-gray-500">Your deployed applications.</p>
      </div>
      <RouterLink :to="{ name: 'new-project' }">
        <Button>
          <Plus :size="16" class="mr-2" />
          New Project
        </Button>
      </RouterLink>
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

    <div v-else-if="error" class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">
      Failed to load projects: {{ error.message }}
    </div>

    <div v-else class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      <RouterLink
        v-for="project in result?.projects"
        :key="project.id"
        :to="{ name: 'project', params: { id: project.id } }"
        class="block"
      >
        <Card class="transition-shadow hover:shadow-md">
          <CardHeader>
            <CardTitle class="text-lg">{{ project.name }}</CardTitle>
            <CardDescription class="flex items-center gap-1">
              <ExternalLink :size="12" />
              {{ project.sourceUrl }}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex flex-wrap gap-2">
              <Badge
                v-for="env in project.environments"
                :key="env.id"
                :variant="syncStatusVariant(env.syncStatus)"
              >
                <GitBranch :size="12" class="mr-1" />
                {{ env.name }}
              </Badge>
            </div>
            <p class="mt-3 text-xs text-gray-400">
              {{ project.environments.length }} environment{{ project.environments.length !== 1 ? 's' : '' }}
            </p>
          </CardContent>
        </Card>
      </RouterLink>
    </div>
  </div>
</template>
