<script setup lang="ts">
import { computed } from 'vue';
import { Rocket, Loader2, CheckCircle, XCircle } from 'lucide-vue-next';
import { useEnvironment } from '@/composables/useEnvironment';
import { useBuild } from '@/composables/useBuild';
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
const build = useBuild();

const envService = computed(() =>
  activeEnvironment.value?.services.find(s => s.name === props.service.name)
);

const hasDeployment = computed(() => !!envService.value?.deployment);

async function handleBuildAndDeploy() {
  const envName = activeEnvironment.value?.name ?? 'development';
  await build.buildAndDeploy(props.projectId, props.service.name, envName);
}

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
</script>

<template>
  <div class="space-y-6">
    <!-- Build & Deploy Action -->
    <div class="flex items-center gap-3">
      <Button
        :disabled="build.isBuilding"
        @click="handleBuildAndDeploy"
      >
        <Loader2
          v-if="build.isBuilding"
          :size="14"
          class="mr-2 animate-spin"
        />
        <Rocket v-else :size="14" class="mr-2" />
        {{ build.isBuilding ? 'Building...' : 'Build & Deploy' }}
      </Button>

      <Badge
        v-if="build.phase"
        :variant="buildPhaseVariant(build.phase)"
        :hide-dot="build.isBuilding"
      >
        <Loader2
          v-if="build.isBuilding"
          :size="12"
          class="mr-1 animate-spin"
        />
        {{ build.phase }}
      </Badge>
    </div>

    <!-- Active Deployment -->
    <div v-if="hasDeployment" class="space-y-3">
      <h3 class="text-sm font-medium text-muted-foreground">Active Deployment</h3>
      <div class="rounded-lg border p-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <component
              :is="envService!.ready ? CheckCircle : XCircle"
              :size="18"
              :class="envService!.ready ? 'text-green-500' : 'text-red-500'"
            />
            <div>
              <p class="text-sm font-medium text-foreground">
                {{ envService!.ready ? 'Online' : 'Not Ready' }}
              </p>
              <p class="text-xs text-muted-foreground">
                {{ envService!.replicas }} replica{{ envService!.replicas !== 1 ? 's' : '' }}
              </p>
            </div>
          </div>
          <Badge variant="outline" class="font-mono text-xs">
            {{ envService!.imageTag }}
          </Badge>
        </div>
      </div>
    </div>

    <!-- No deployment -->
    <EmptyState
      v-else
      title="No deployment"
      description="This service hasn't been deployed to this environment yet. Click Build & Deploy to get started."
      pattern="diagonal"
    />
  </div>
</template>
