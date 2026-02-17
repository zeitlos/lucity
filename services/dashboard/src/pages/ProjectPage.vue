<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, RouterLink } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { ArrowLeft, GitBranch, Container, Globe, Lock, Layers } from 'lucide-vue-next';
import { ProjectQuery } from '@/graphql/projects';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Separator } from '@/components/ui/separator';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableEmpty,
} from '@/components/ui/table';
import EmptyState from '@/components/EmptyState.vue';

const route = useRoute();
const projectId = computed(() => route.params.id as string);

const { result, loading, error } = useQuery(ProjectQuery, () => ({
  id: projectId.value,
}));

const project = computed(() => result.value?.project);

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
    <div v-if="loading" class="space-y-6">
      <Skeleton class="h-8 w-48" />
      <Skeleton class="h-4 w-64" />
      <div class="grid gap-4 md:grid-cols-2">
        <Skeleton class="h-32" />
        <Skeleton class="h-32" />
      </div>
    </div>

    <div v-else-if="error" class="rounded-lg border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">
      Failed to load project: {{ error.message }}
    </div>

    <template v-else-if="project">
      <div class="mb-6">
        <RouterLink
          :to="{ name: 'projects' }"
          class="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft :size="14" />
          Projects
        </RouterLink>
        <h1 class="text-2xl font-semibold text-foreground">{{ project.name }}</h1>
        <p class="mt-1 text-sm text-muted-foreground">{{ project.sourceUrl }}</p>
      </div>

      <div class="space-y-8">
        <section>
          <h2 class="mb-4 text-lg font-medium text-foreground">Environments</h2>
          <EmptyState
            v-if="project.environments.length === 0"
            :icon="GitBranch"
            title="No environments"
            description="Environments will appear here once the project is deployed."
          />
          <div v-else class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <RouterLink
              v-for="env in project.environments"
              :key="env.id"
              :to="{ name: 'environment', params: { id: project.id, env: env.name } }"
              class="block"
            >
              <Card class="transition-shadow hover:shadow-md">
                <CardHeader class="pb-3">
                  <div class="flex items-center justify-between">
                    <CardTitle class="text-base">
                      <div class="flex items-center gap-2">
                        <GitBranch :size="16" />
                        {{ env.name }}
                      </div>
                    </CardTitle>
                    <Badge :variant="syncStatusVariant(env.syncStatus)">
                      {{ env.syncStatus }}
                    </Badge>
                  </div>
                  <CardDescription>{{ env.namespace }}</CardDescription>
                </CardHeader>
                <CardContent>
                  <p class="text-xs text-muted-foreground">
                    {{ env.services.length }} service{{ env.services.length !== 1 ? 's' : '' }} deployed
                  </p>
                </CardContent>
              </Card>
            </RouterLink>
          </div>
        </section>

        <Separator />

        <section>
          <h2 class="mb-4 text-lg font-medium text-foreground">Services</h2>
          <Card>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Image</TableHead>
                  <TableHead>Port</TableHead>
                  <TableHead>Visibility</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <template v-if="project.services.length === 0">
                  <TableEmpty :colspan="4">
                    <div class="flex flex-col items-center py-6">
                      <Layers :size="24" class="mb-2 text-muted-foreground" />
                      <p>No services configured yet.</p>
                      <p class="mt-1 text-xs">Services will appear once detected from your repository.</p>
                    </div>
                  </TableEmpty>
                </template>
                <template v-else>
                  <TableRow v-for="svc in project.services" :key="svc.name">
                    <TableCell class="font-medium">
                      <div class="flex items-center gap-2">
                        <Container :size="14" />
                        {{ svc.name }}
                      </div>
                    </TableCell>
                    <TableCell class="font-mono text-sm text-muted-foreground">{{ svc.image }}</TableCell>
                    <TableCell>{{ svc.port || '—' }}</TableCell>
                    <TableCell>
                      <Badge :variant="svc.public ? 'default' : 'secondary'">
                        <component :is="svc.public ? Globe : Lock" :size="12" class="mr-1" />
                        {{ svc.public ? 'Public' : 'Private' }}
                      </Badge>
                    </TableCell>
                  </TableRow>
                </template>
              </TableBody>
            </Table>
          </Card>
        </section>
      </div>
    </template>
  </div>
</template>
