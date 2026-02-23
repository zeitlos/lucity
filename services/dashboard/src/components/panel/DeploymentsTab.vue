<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';
import { Rocket, Loader2, Check, Circle, ChevronRight, AlertCircle, TriangleAlert, Terminal, ExternalLink, GitCommitHorizontal, RotateCcw, RefreshCw } from 'lucide-vue-next';
import { useEnvironment } from '@/composables/useEnvironment';
import { useDeploy } from '@/composables/useDeploy';
import { useDeploymentLogsPanel } from '@/composables/useDeploymentLogsPanel';
import { apolloClient } from '@/lib/apollo';
import { ActiveDeploymentQuery, RollbackMutation } from '@/graphql/services';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import EmptyState from '@/components/EmptyState.vue';
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  projectId: string;
  service: {
    name: string;
    image: string;
    port: number;
    framework?: string;
  };
}>();

const { activeEnvironment } = useEnvironment();
const deploy = useDeploy();
const logsPanel = useDeploymentLogsPanel();

function showLogs() {
  if (deploy.deployId) {
    logsPanel.open(deploy.deployId, props.service.name);
  }
}

onMounted(async () => {
  const envName = activeEnvironment.value?.name;
  if (!envName) return;

  try {
    const { data } = await apolloClient.query({
      query: ActiveDeploymentQuery,
      variables: { projectId: props.projectId, service: props.service.name, environment: envName },
      fetchPolicy: 'network-only',
    });

    const active = data?.activeDeployment;
    if (active?.id) {
      deploy.pollDeploy(active.id);
      deploy.phase = active.phase;
      deploy.rolloutHealth = active.rolloutHealth ?? null;
      deploy.rolloutMessage = active.rolloutMessage ?? null;
    }
  } catch {
    // No active deployment — nothing to resume.
  }
});

const envService = computed(() =>
  activeEnvironment.value?.services.find(s => s.name === props.service.name)
);

const deployments = computed(() => envService.value?.deployments ?? []);
const hasDeployments = computed(() => deployments.value.length > 0);

const expandedId = ref<string | null>(null);

function toggleExpanded(id: string) {
  expandedId.value = expandedId.value === id ? null : id;
}

async function handleDeploy() {
  const envName = activeEnvironment.value?.name ?? 'development';
  await deploy.startDeploy(props.projectId, props.service.name, envName);
}

async function handleRollback(imageTag: string) {
  const envName = activeEnvironment.value?.name;
  if (!envName) return;

  try {
    await apolloClient.mutate({
      mutation: RollbackMutation,
      variables: {
        input: {
          projectId: props.projectId,
          service: props.service.name,
          environment: envName,
          imageTag,
        },
      },
    });
    toast.success('Rollback initiated', { description: `Rolling back to ${imageTag}` });
  } catch (e: unknown) {
    toast.error('Rollback failed', { description: errorMessage(e) });
  }
}

async function handleRedeploy(imageTag: string) {
  const envName = activeEnvironment.value?.name ?? 'development';
  await deploy.startDeploy(props.projectId, props.service.name, envName, imageTag);
}

// Deploy pipeline stages
const STAGES = ['Initializing', 'Building', 'Deploying'] as const;

// Map DeployPhase enum to stage index (0-based, -1 = not started)
function phaseToStageIndex(phase: string): number {
  switch (phase) {
    case 'QUEUED':
    case 'CLONING':
      return 0;
    case 'BUILDING':
    case 'PUSHING':
      return 1;
    case 'DEPLOYING':
      return 2;
    case 'SUCCEEDED':
      return 3; // all complete
    default:
      return -1;
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
        :variant="deploy.phase === 'SUCCEEDED' ? 'default' : deploy.phase === 'FAILED' ? 'destructive' : 'secondary'"
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

    <!-- In-flight deploy stages -->
    <div
      v-if="deploy.isDeploying && deploy.phase"
      class="rounded-lg border border-border/60 bg-muted/30"
    >
      <div class="px-3 py-2.5">
        <div class="space-y-1">
          <div
            v-for="(stage, idx) in STAGES"
            :key="stage"
            class="flex items-center gap-2.5 py-1"
          >
            <!-- Stage indicator -->
            <div class="flex h-4 w-4 items-center justify-center">
              <Check
                v-if="phaseToStageIndex(deploy.phase!) > idx"
                :size="14"
                class="text-[var(--status-ok)]"
              />
              <Loader2
                v-else-if="!deploy.error && phaseToStageIndex(deploy.phase!) === idx"
                :size="14"
                class="animate-spin text-[var(--primary)]"
              />
              <AlertCircle
                v-else-if="deploy.error"
                :size="14"
                class="text-[var(--status-danger)]"
              />
              <Circle
                v-else
                :size="8"
                class="text-muted-foreground/40"
              />
            </div>
            <span
              class="text-xs"
              :class="phaseToStageIndex(deploy.phase!) >= idx
                ? 'text-foreground'
                : 'text-muted-foreground/60'"
            >
              {{ stage }}
            </span>
          </div>
        </div>

        <!-- Rollout health status during DEPLOYING phase -->
        <div
          v-if="deploy.phase === 'DEPLOYING' && deploy.rolloutHealth"
          class="mt-2 rounded-md px-2.5 py-2"
          :class="deploy.rolloutHealth === 'DEGRADED'
            ? 'bg-[var(--status-danger)]/10'
            : 'bg-muted/50'"
        >
          <div class="flex items-start gap-2">
            <TriangleAlert
              v-if="deploy.rolloutHealth === 'DEGRADED'"
              :size="13"
              class="mt-0.5 shrink-0 text-[var(--status-danger)]"
            />
            <Loader2
              v-else-if="deploy.rolloutHealth === 'PROGRESSING'"
              :size="13"
              class="mt-0.5 shrink-0 animate-spin text-[var(--status-warn)]"
            />
            <div class="min-w-0 space-y-0.5">
              <p class="text-xs font-medium text-foreground">
                {{ deploy.rolloutHealth === 'DEGRADED' ? 'Rollout degraded' : 'Waiting for pods' }}
              </p>
              <p
                v-if="deploy.rolloutMessage"
                class="break-words font-mono text-[11px] text-muted-foreground"
              >
                {{ deploy.rolloutMessage }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Show Logs button -->
      <div class="border-t border-border/40 px-3 py-2">
        <Button
          variant="ghost"
          size="sm"
          class="h-7 text-xs text-muted-foreground"
          @click="showLogs"
        >
          <Terminal :size="13" class="mr-1.5" />
          Show Logs
        </Button>
      </div>
    </div>

    <!-- Deploy error (FAILED phase) -->
    <div
      v-if="deploy.phase === 'FAILED' && (deploy.error || deploy.rolloutMessage)"
      class="rounded-lg border border-[var(--status-danger)]/30 bg-[var(--status-danger)]/5 px-3 py-2.5"
    >
      <div class="flex items-start gap-2">
        <AlertCircle
          :size="14"
          class="mt-0.5 shrink-0 text-[var(--status-danger)]"
        />
        <div class="min-w-0 space-y-0.5">
          <p class="text-xs font-medium text-[var(--status-danger)]">Deploy failed</p>
          <p class="break-words font-mono text-[11px] text-muted-foreground">
            {{ deploy.error || deploy.rolloutMessage }}
          </p>
          <button
            class="mt-1 text-[11px] text-muted-foreground underline decoration-muted-foreground/40 underline-offset-2 hover:text-foreground"
            @click="showLogs"
          >
            Show Logs
          </button>
        </div>
      </div>
    </div>

    <!-- Deployment History -->
    <div v-if="hasDeployments" class="space-y-3">
      <h3 class="text-sm font-medium text-muted-foreground">Deployment History</h3>
      <div class="space-y-2">
        <Collapsible
          v-for="dep in deployments"
          :key="dep.id"
          :open="expandedId === dep.id"
        >
          <CollapsibleTrigger
            class="w-full cursor-pointer"
            @click="toggleExpanded(dep.id)"
          >
            <div
              class="rounded-lg border px-3 py-2.5 text-left transition-colors hover:bg-muted/50"
              :class="dep.active ? 'border-[var(--primary)]/30 border-l-2 border-l-[var(--primary)]' : 'border-border/60'"
            >
              <div class="flex items-center justify-between gap-2">
                <div class="flex min-w-0 items-center gap-2">
                  <ChevronRight
                    :size="14"
                    class="shrink-0 text-muted-foreground transition-transform"
                    :class="expandedId === dep.id ? 'rotate-90' : ''"
                  />
                  <Badge
                    v-if="dep.active"
                    variant="default"
                    class="shrink-0 text-xs"
                  >
                    Active
                  </Badge>
                  <span
                    v-if="dep.sourceCommitMessage"
                    class="truncate text-xs text-foreground"
                    :title="dep.sourceCommitMessage"
                  >
                    {{ dep.sourceCommitMessage }}
                  </span>
                  <Badge variant="outline" class="shrink-0 font-mono text-xs">
                    <GitCommitHorizontal :size="10" class="shrink-0" />
                    {{ dep.imageTag }}
                  </Badge>
                </div>
                <div class="flex shrink-0 items-center gap-2">
                  <a
                    v-if="dep.sourceUrl"
                    :href="dep.sourceUrl"
                    target="_blank"
                    rel="noopener"
                    class="text-muted-foreground hover:text-foreground"
                    title="View commit on GitHub"
                    @click.stop
                  >
                    <ExternalLink :size="12" />
                  </a>
                  <span
                    v-if="dep.timestamp"
                    class="text-xs text-muted-foreground"
                  >
                    {{ formatRelativeTime(dep.timestamp) }}
                  </span>
                </div>
              </div>
            </div>
          </CollapsibleTrigger>

          <CollapsibleContent>
            <div class="ml-3 border-l border-border/40 py-2 pl-4">
              <!-- Stages (historical — all succeeded) -->
              <div class="space-y-1">
                <div
                  v-for="stage in STAGES"
                  :key="stage"
                  class="flex items-center gap-2.5 py-0.5"
                >
                  <Check :size="12" class="text-[var(--status-ok)]" />
                  <span class="text-xs text-foreground">{{ stage }}</span>
                </div>
              </div>

              <!-- Details -->
              <div class="mt-3 space-y-1 text-xs text-muted-foreground">
                <p v-if="dep.active && envService">
                  {{ envService.replicas }} replica{{ envService.replicas !== 1 ? 's' : '' }}
                  <template v-if="envService.ready"> &middot; healthy</template>
                  <template v-else> &middot; not ready</template>
                </p>
                <p v-if="dep.message" class="font-mono">
                  {{ dep.message }}
                </p>
              </div>

              <!-- Actions -->
              <div class="mt-3 flex gap-2">
                <Button
                  v-if="!dep.active"
                  variant="outline"
                  size="sm"
                  class="h-7 text-xs"
                  :disabled="deploy.isDeploying"
                  @click.stop="handleRollback(dep.imageTag)"
                >
                  <RotateCcw :size="12" class="mr-1.5" />
                  Rollback
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  class="h-7 text-xs"
                  :disabled="deploy.isDeploying"
                  @click.stop="handleRedeploy(dep.imageTag)"
                >
                  <RefreshCw :size="12" class="mr-1.5" />
                  Redeploy
                </Button>
              </div>
            </div>
          </CollapsibleContent>
        </Collapsible>
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
