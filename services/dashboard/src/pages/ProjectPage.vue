<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, RouterLink } from 'vue-router';
import { useQuery, useMutation, useLazyQuery } from '@vue/apollo-composable';
import {
  ArrowLeft,
  GitBranch,
  Globe,
  Lock,
  Layers,
  Rocket,
  Trash2,
  Plus,
  Scan,
  Loader2,
} from 'lucide-vue-next';
import { ProjectQuery } from '@/graphql/projects';
import {
  DetectServicesQuery,
  AddServiceMutation,
  RemoveServiceMutation,
} from '@/graphql/services';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast } from '@/components/ui/sonner';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableEmpty,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
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
import EmptyState from '@/components/EmptyState.vue';
import FrameworkIcon from '@/components/FrameworkIcon.vue';
import { errorMessage } from '@/lib/utils';
import { useBuild, type BuildState } from '@/composables/useBuild';

const route = useRoute();
const projectId = computed(() => route.params.id as string);

const { result, loading, error, refetch } = useQuery(ProjectQuery, () => ({
  id: projectId.value,
}));

const project = computed(() => result.value?.project);

// Detection
const {
  load: loadDetection,
  result: detectResult,
  loading: detecting,
  error: detectError,
} = useLazyQuery(DetectServicesQuery, () => ({
  projectId: projectId.value,
}), {
  fetchPolicy: 'network-only',
});

const detectedServices = computed(() => detectResult.value?.detectServices ?? []);
const showDetectionPanel = ref(false);

// Auto-detect when project has no services
watch(
  () => project.value,
  (proj) => {
    if (proj && proj.services.length === 0 && !detectResult.value) {
      showDetectionPanel.value = true;
      loadDetection();
    }
  },
  { immediate: true },
);

// Add Service
const { mutate: addServiceMutate, loading: addingService } = useMutation(AddServiceMutation);

async function confirmDetectedService(detected: {
  name: string;
  framework: string;
  suggestedPort: number;
}) {
  try {
    const res = await addServiceMutate({
      input: {
        projectId: projectId.value,
        name: detected.name,
        port: detected.suggestedPort,
        public: true,
        framework: detected.framework || undefined,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to add service', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Service added', { description: `${detected.name} configured` });
    showDetectionPanel.value = false;
    refetch();
  } catch (e: unknown) {
    toast.error('Failed to add service', { description: errorMessage(e) });
  }
}

// Manual add service dialog
const addDialogOpen = ref(false);
const newServiceName = ref('web');
const newServicePort = ref(3000);
const newServicePublic = ref(true);

async function handleAddService() {
  try {
    const res = await addServiceMutate({
      input: {
        projectId: projectId.value,
        name: newServiceName.value,
        port: newServicePort.value,
        public: newServicePublic.value,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to add service', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Service added');
    addDialogOpen.value = false;
    newServiceName.value = 'web';
    newServicePort.value = 3000;
    newServicePublic.value = true;
    refetch();
  } catch (e: unknown) {
    toast.error('Failed to add service', { description: errorMessage(e) });
  }
}

// Remove service
const { mutate: removeServiceMutate } = useMutation(RemoveServiceMutation);

async function handleRemoveService(service: string) {
  try {
    const res = await removeServiceMutate({
      projectId: projectId.value,
      service,
    });

    if (res?.errors?.length) {
      toast.error('Failed to remove service', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Service removed');
    refetch();
  } catch (e: unknown) {
    toast.error('Failed to remove service', { description: errorMessage(e) });
  }
}

// Build
const builds = ref<Record<string, BuildState>>({});

function getBuild(service: string) {
  if (!builds.value[service]) {
    builds.value[service] = useBuild();
  }
  return builds.value[service];
}

async function handleBuildAndDeploy(service: string) {
  const build = getBuild(service);
  await build.buildAndDeploy(projectId.value, service, 'development');
}

// Refetch project when any build reaches DEPLOYED to update env sync status
watch(
  () => Object.values(builds.value).map(b => b.phase),
  (phases) => {
    if (phases.some(p => p === 'DEPLOYED')) {
      refetch();
    }
  },
);

function buildPhaseVariant(phase: string) {
  switch (phase) {
    case 'DEPLOYED': return 'default';
    case 'SUCCEEDED': return 'default';
    case 'FAILED': return 'destructive';
    case 'BUILDING': return 'secondary';
    case 'PUSHING': return 'secondary';
    case 'DEPLOYING': return 'secondary';
    default: return 'outline';
  }
}

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

    <div
      v-else-if="error"
      class="rounded-lg border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive"
    >
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
        <!-- Detection Panel -->
        <section v-if="showDetectionPanel">
          <Card class="border-primary/30">
            <CardHeader>
              <CardTitle class="flex items-center gap-2 text-base">
                <Scan :size="18" />
                Service Detection
              </CardTitle>
              <CardDescription>
                We scanned your repository to detect services.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div v-if="detecting" class="flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 :size="16" class="animate-spin" />
                Scanning repository...
              </div>

              <div
                v-else-if="detectError"
                class="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive"
              >
                <p class="font-medium">Detection failed</p>
                <p class="mt-1 text-xs">{{ detectError.message }}</p>
              </div>

              <div v-else-if="detectedServices.length > 0" class="space-y-4">
                <div
                  v-for="detected in detectedServices"
                  :key="detected.name"
                  class="flex items-center justify-between rounded-lg border p-4"
                >
                  <div class="flex items-center gap-3">
                    <FrameworkIcon :framework="detected.framework" :size="24" />
                    <div>
                      <p class="font-medium">
                        {{ detected.framework || detected.provider }} app detected
                      </p>
                      <p class="text-sm text-muted-foreground">
                        Port {{ detected.suggestedPort }}
                        <span v-if="detected.startCommand">
                          &middot; <code class="text-xs">{{ detected.startCommand }}</code>
                        </span>
                      </p>
                    </div>
                  </div>
                  <Button
                    size="sm"
                    :disabled="addingService"
                    @click="confirmDetectedService(detected)"
                  >
                    {{ addingService ? 'Adding...' : 'Confirm & Add' }}
                  </Button>
                </div>
              </div>

              <div v-else class="text-sm text-muted-foreground">
                No services detected. You can add one manually.
              </div>

              <div class="mt-4 flex gap-2">
                <Button
                  v-if="detectError"
                  variant="outline"
                  size="sm"
                  @click="loadDetection(undefined, { fetchPolicy: 'network-only' })"
                >
                  Retry
                </Button>
                <Button
                  v-if="detectedServices.length === 0 || detectError"
                  variant="outline"
                  size="sm"
                  @click="showDetectionPanel = false"
                >
                  Dismiss
                </Button>
              </div>
            </CardContent>
          </Card>
        </section>

        <!-- Environments -->
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
                    {{ env.services.length }} service{{ env.services.length !== 1 ? 's' : '' }}
                    deployed
                  </p>
                </CardContent>
              </Card>
            </RouterLink>
          </div>
        </section>

        <Separator />

        <!-- Services -->
        <section>
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-lg font-medium text-foreground">Services</h2>
            <div class="flex gap-2">
              <Button
                v-if="!showDetectionPanel"
                variant="outline"
                size="sm"
                @click="showDetectionPanel = true; loadDetection();"
              >
                <Scan :size="14" class="mr-1" />
                Detect
              </Button>

              <Dialog v-model:open="addDialogOpen">
                <DialogTrigger as-child>
                  <Button variant="outline" size="sm">
                    <Plus :size="14" class="mr-1" />
                    Add Service
                  </Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Add Service</DialogTitle>
                    <DialogDescription>
                      Manually configure a service for this project.
                    </DialogDescription>
                  </DialogHeader>
                  <div class="space-y-4 py-4">
                    <div class="space-y-2">
                      <Label for="svc-name">Name</Label>
                      <Input
                        id="svc-name"
                        v-model="newServiceName"
                        placeholder="web"
                      />
                    </div>
                    <div class="space-y-2">
                      <Label for="svc-port">Port</Label>
                      <Input
                        id="svc-port"
                        v-model.number="newServicePort"
                        type="number"
                        placeholder="3000"
                      />
                    </div>
                    <div class="flex items-center gap-2">
                      <Switch
                        id="svc-public"
                        :checked="newServicePublic"
                        @update:checked="newServicePublic = $event"
                      />
                      <Label for="svc-public">Public</Label>
                    </div>
                  </div>
                  <DialogFooter>
                    <Button
                      :disabled="addingService || !newServiceName"
                      @click="handleAddService"
                    >
                      {{ addingService ? 'Adding...' : 'Add Service' }}
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </div>
          </div>

          <Card>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Image</TableHead>
                  <TableHead>Port</TableHead>
                  <TableHead>Visibility</TableHead>
                  <TableHead class="w-[180px]">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <template v-if="project.services.length === 0">
                  <TableEmpty :colspan="5">
                    <div class="flex flex-col items-center py-6">
                      <Layers :size="24" class="mb-2 text-muted-foreground" />
                      <p>No services configured yet.</p>
                      <p class="mt-1 text-xs">
                        Services will appear once detected from your repository.
                      </p>
                    </div>
                  </TableEmpty>
                </template>
                <template v-else>
                  <TableRow v-for="svc in project.services" :key="svc.name">
                    <TableCell class="font-medium">
                      <div class="flex items-center gap-2">
                        <FrameworkIcon :framework="svc.framework" :size="16" />
                        {{ svc.name }}
                      </div>
                    </TableCell>
                    <TableCell class="max-w-[200px] truncate font-mono text-sm text-muted-foreground">
                      {{ svc.image }}
                    </TableCell>
                    <TableCell>{{ svc.port || '—' }}</TableCell>
                    <TableCell>
                      <Badge :variant="svc.public ? 'default' : 'secondary'">
                        <component
                          :is="svc.public ? Globe : Lock"
                          :size="12"
                          class="mr-1"
                        />
                        {{ svc.public ? 'Public' : 'Private' }}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div class="flex items-center gap-1">
                        <!-- Build phase badge -->
                        <Badge
                          v-if="builds[svc.name]?.phase"
                          :variant="buildPhaseVariant(builds[svc.name].phase!)"
                          class="mr-1"
                        >
                          <Loader2
                            v-if="builds[svc.name]?.isBuilding"
                            :size="12"
                            class="mr-1 animate-spin"
                          />
                          {{ builds[svc.name].phase }}
                        </Badge>

                        <!-- Build & Deploy button -->
                        <Button
                          variant="outline"
                          size="icon"
                          class="h-7 w-7"
                          title="Build & Deploy"
                          :disabled="builds[svc.name]?.isBuilding"
                          @click="handleBuildAndDeploy(svc.name)"
                        >
                          <Rocket :size="14" />
                        </Button>

                        <!-- Remove button -->
                        <AlertDialog>
                          <AlertDialogTrigger as-child>
                            <Button
                              variant="ghost"
                              size="icon"
                              class="h-7 w-7 text-muted-foreground hover:text-destructive"
                            >
                              <Trash2 :size="14" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Remove service</AlertDialogTitle>
                              <AlertDialogDescription>
                                This will remove <strong>{{ svc.name }}</strong> from the project
                                configuration. This action cannot be undone.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction @click="handleRemoveService(svc.name)">
                                Remove
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
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
