<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, RouterLink } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { ArrowLeft, GitBranch, CheckCircle, XCircle, Container } from 'lucide-vue-next';
import { ProjectQuery } from '@/graphql/projects';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableEmpty,
} from '@/components/ui/table';

const route = useRoute();
const projectId = computed(() => route.params.id as string);
const envName = computed(() => route.params.env as string);

const { result, loading, error } = useQuery(ProjectQuery, () => ({
  id: projectId.value,
}));

const project = computed(() => result.value?.project);
const environment = computed(() =>
  project.value?.environments.find((e: { name: string }) => e.name === envName.value)
);

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
      <Skeleton class="h-48 w-full" />
    </div>

    <div v-else-if="error" class="rounded-lg border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">
      Failed to load environment: {{ error.message }}
    </div>

    <template v-else-if="project && environment">
      <div class="mb-6">
        <RouterLink
          :to="{ name: 'project', params: { id: project.id } }"
          class="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft :size="14" />
          {{ project.name }}
        </RouterLink>

        <div class="flex items-center gap-3">
          <GitBranch :size="20" class="text-muted-foreground" />
          <h1 class="text-2xl font-semibold text-foreground">{{ environment.name }}</h1>
          <Badge :variant="syncStatusVariant(environment.syncStatus)">
            {{ environment.syncStatus }}
          </Badge>
        </div>
        <p class="mt-1 text-sm text-muted-foreground">
          Namespace: <span class="font-mono">{{ environment.namespace }}</span>
        </p>
      </div>

      <section>
        <h2 class="mb-4 text-lg font-medium text-foreground">Deployed Services</h2>
        <Card>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Service</TableHead>
                <TableHead>Image Tag</TableHead>
                <TableHead>Replicas</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="environment.services.length === 0">
                <TableEmpty :colspan="4">
                  <div class="flex flex-col items-center py-6">
                    <Container :size="24" class="mb-2 text-muted-foreground" />
                    <p>No services deployed yet.</p>
                    <p class="mt-1 text-xs">Services will appear here after the first deployment.</p>
                  </div>
                </TableEmpty>
              </template>
              <template v-else>
                <TableRow v-for="svc in environment.services" :key="svc.name">
                  <TableCell class="font-medium">{{ svc.name }}</TableCell>
                  <TableCell>
                    <Badge variant="outline" class="font-mono">{{ svc.imageTag }}</Badge>
                  </TableCell>
                  <TableCell>{{ svc.replicas }}</TableCell>
                  <TableCell>
                    <div class="flex items-center gap-2">
                      <component
                        :is="svc.ready ? CheckCircle : XCircle"
                        :size="16"
                        :class="svc.ready ? 'text-green-500' : 'text-red-500'"
                      />
                      {{ svc.ready ? 'Ready' : 'Not Ready' }}
                    </div>
                  </TableCell>
                </TableRow>
              </template>
            </TableBody>
          </Table>
        </Card>
      </section>
    </template>
  </div>
</template>
