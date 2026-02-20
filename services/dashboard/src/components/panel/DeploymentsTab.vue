<script setup lang="ts">
import { computed } from 'vue';
import { Rocket, Loader2, CheckCircle, XCircle, Clock } from 'lucide-vue-next';
import { useEnvironment } from '@/composables/useEnvironment';
import { useDeploy } from '@/composables/useDeploy';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import EmptyState from '@/components/EmptyState.vue';

const props = defineProps<{
  projectId: string;
  service: {
    name: string;
    image: string;
    port: number;
    public: boolean;
    framework?: string;
  };
}>();

const { activeEnvironment } = useEnvironment();
const deploy = useDeploy();

const envService = computed(() =>
  activeEnvironment.value?.services.find(s => s.name === props.service.name)
);

const deployments = computed(() => envService.value?.deployments ?? []);
const hasDeployments = computed(() => deployments.value.length > 0);

async function handleDeploy() {
  const envName = activeEnvironment.value?.name ?? 'development';
  await deploy.startDeploy(props.projectId, props.service.name, envName);
}

function phaseVariant(phase: string) {
  switch (phase) {
    case 'SUCCEEDED': return 'default';
    case 'FAILED': return 'destructive';
    case 'BUILDING': return 'secondary';
    case 'PUSHING': return 'secondary';
    case 'DEPLOYING': return 'secondary';
    default: return 'outline';
  }
}

function formatRelativeTime(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 60) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 30) return `${diffDays}d ago`;
  return date.toLocaleDateString();
}
</script>

<template>
  <div class="space-y-6">
    <!-- Deploy Action -->
    <div class="flex items-center gap-3">
      <Button
        :disabled="deploy.isDeploying"
        @click="handleDeploy"
      >
        <Loader2
          v-if="deploy.isDeploying"
          :size="14"
          class="mr-2 animate-spin"
        />
        <Rocket v-else :size="14" class="mr-2" />
        {{ deploy.isDeploying ? 'Deploying...' : 'Deploy' }}
      </Button>

      <Badge
        v-if="deploy.phase"
        :variant="phaseVariant(deploy.phase)"
        :hide-dot="deploy.isDeploying"
      >
        <Loader2
          v-if="deploy.isDeploying"
          :size="12"
          class="mr-1 animate-spin"
        />
        {{ deploy.phase }}
      </Badge>
    </div>

    <!-- Deployment History -->
    <div v-if="hasDeployments" class="space-y-3">
      <h3 class="text-sm font-medium text-muted-foreground">Deployment History</h3>
      <div class="space-y-2">
        <div
          v-for="dep in deployments"
          :key="dep.id"
          class="rounded-lg border p-3 transition-colors"
          :class="dep.active ? 'border-green-500/30 bg-green-500/5' : ''"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <component
                v-if="dep.active"
                :is="envService!.ready ? CheckCircle : XCircle"
                :size="14"
                :class="envService!.ready ? 'text-green-500' : 'text-red-500'"
              />
              <Clock
                v-else
                :size="14"
                class="text-muted-foreground"
              />
              <Badge
                v-if="dep.active"
                variant="default"
                class="text-xs"
              >
                Active
              </Badge>
              <Badge variant="outline" class="font-mono text-xs">
                {{ dep.imageTag }}
              </Badge>
            </div>
            <span
              v-if="dep.timestamp"
              class="text-xs text-muted-foreground"
            >
              {{ formatRelativeTime(dep.timestamp) }}
            </span>
          </div>
          <div v-if="dep.active && envService" class="mt-1 pl-6">
            <p class="text-xs text-muted-foreground">
              {{ envService.replicas }} replica{{ envService.replicas !== 1 ? 's' : '' }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- No deployment -->
    <EmptyState
      v-else-if="!deploy.isDeploying"
      title="No deployment"
      description="This service hasn't been deployed to this environment yet. Click Deploy to get started."
      pattern="diagonal"
    />
  </div>
</template>
