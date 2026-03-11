<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';
import {
  Rocket, Loader2, Check, AlertCircle, TriangleAlert, Terminal,
  ExternalLink, GitCommitHorizontal, RotateCcw, RefreshCw,
  MoreVertical, ChevronDown, Circle, Container,
} from 'lucide-vue-next';
import { useEnvironment } from '@/composables/useEnvironment';
import { useDeploy } from '@/composables/useDeploy';
import { useDeploymentLogsPanel } from '@/composables/useDeploymentLogsPanel';
import { apolloClient } from '@/lib/apollo';
import { ActiveDeploymentQuery, RollbackMutation } from '@/graphql/services';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
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
    sourceUrl?: string;
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

const isImageBased = computed(() => !props.service.sourceUrl);

const envService = computed(() =>
  activeEnvironment.value?.services.find(s => s.name === props.service.name)
);

const deployments = computed(() => envService.value?.deployments ?? []);
const hasDeployments = computed(() => deployments.value.length > 0);
const activeDeployment = computed(() => deployments.value.find(d => d.active));
const pastDeployments = computed(() => deployments.value.filter(d => !d.active));

const showActiveDetails = ref(false);

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
      return 3;
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

function deployLabel(dep: { sourceCommitMessage?: string; imageTag: string }): string {
  return dep.sourceCommitMessage || dep.imageTag;
}
</script>

<template>
  <div class="space-y-4">
    <!-- Image-based service info -->
    <div
      v-if="isImageBased"
      class="flex items-start gap-2 rounded-lg border border-border/60 bg-muted/30 px-3 py-2.5"
    >
      <Container :size="14" class="mt-0.5 shrink-0 text-muted-foreground" />
      <div class="min-w-0 space-y-0.5">
        <p class="text-sm font-medium text-foreground">External container image</p>
        <p class="text-xs text-muted-foreground">
          This service uses a pre-built image. Deployments sync automatically via ArgoCD.
        </p>
      </div>
    </div>

    <!-- Deploy Action (source-based services only) -->
    <div v-else class="flex items-center gap-3">
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

        <!-- Rollout health status -->
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

    <!-- Deploy error -->
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

    <!-- Active Deployment Card -->
    <div v-if="activeDeployment" class="space-y-0">
      <div class="rounded-lg border border-border/60 bg-card">
        <!-- Main row -->
        <div class="flex items-start gap-3 px-4 py-3">
          <Badge
            variant="default"
            class="mt-0.5 shrink-0"
          >
            Active
          </Badge>

          <div class="min-w-0 flex-1">
            <p
              class="truncate text-sm font-medium text-foreground"
              :title="activeDeployment.sourceCommitMessage || activeDeployment.imageTag"
            >
              {{ deployLabel(activeDeployment) }}
            </p>
            <div class="mt-0.5 flex items-center gap-1.5 text-xs text-muted-foreground">
              <span v-if="activeDeployment.timestamp">{{ formatRelativeTime(activeDeployment.timestamp) }}</span>
              <span v-if="activeDeployment.sourceUrl">via GitHub</span>
            </div>
          </div>

          <div class="flex shrink-0 items-center gap-1.5">
            <Button
              v-if="deploy.deployId"
              variant="outline"
              size="sm"
              class="h-8 text-xs"
              @click="showLogs"
            >
              View logs
            </Button>

            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button
                  variant="ghost"
                  size="sm"
                  class="h-8 w-8 p-0"
                >
                  <MoreVertical :size="16" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem @click="handleRedeploy(activeDeployment.imageTag)">
                  <RefreshCw :size="14" class="mr-2" />
                  Redeploy
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="activeDeployment.sourceUrl" />
                <DropdownMenuItem v-if="activeDeployment.sourceUrl" as-child>
                  <a
                    :href="activeDeployment.sourceUrl"
                    target="_blank"
                    rel="noopener"
                  >
                    <ExternalLink :size="14" class="mr-2" />
                    View on GitHub
                  </a>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        <!-- Deployment successful expandable -->
        <Collapsible v-model:open="showActiveDetails">
          <CollapsibleTrigger class="flex w-full cursor-pointer items-center gap-2 border-t border-border/40 px-4 py-2.5 text-left">
            <template v-if="envService?.ready">
              <Check :size="14" class="shrink-0 text-[var(--status-ok)]" />
              <span class="flex-1 text-xs font-medium text-[var(--status-ok)]">Deployment successful</span>
            </template>
            <template v-else>
              <Loader2 :size="14" class="shrink-0 animate-spin text-[var(--status-warn)]" />
              <span class="flex-1 text-xs font-medium text-[var(--status-warn)]">Waiting for pods</span>
            </template>
            <ChevronDown
              :size="14"
              class="shrink-0 text-muted-foreground transition-transform"
              :class="showActiveDetails ? 'rotate-180' : ''"
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="space-y-2 border-t border-border/40 px-4 py-3">
              <div class="flex items-center gap-2">
                <GitCommitHorizontal :size="12" class="shrink-0 text-muted-foreground" />
                <span class="font-mono text-xs text-muted-foreground">{{ activeDeployment.imageTag }}</span>
              </div>
              <div v-if="envService" class="text-xs text-muted-foreground">
                {{ envService.replicas }} replica{{ envService.replicas !== 1 ? 's' : '' }}
                <template v-if="envService.ready"> &middot; healthy</template>
                <template v-else> &middot; not ready</template>
              </div>
            </div>
          </CollapsibleContent>
        </Collapsible>
      </div>
    </div>

    <!-- History Section -->
    <div v-if="pastDeployments.length > 0" class="space-y-2">
      <h3 class="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        History
      </h3>

      <div class="space-y-2">
        <div
          v-for="dep in pastDeployments"
          :key="dep.id"
          class="flex items-start gap-3 rounded-lg border border-border/60 bg-muted/30 px-4 py-3"
        >
          <div class="min-w-0 flex-1">
            <p
              class="truncate text-sm text-foreground"
              :title="dep.sourceCommitMessage || dep.imageTag"
            >
              {{ deployLabel(dep) }}
            </p>
            <div class="mt-0.5 flex items-center gap-1.5 text-xs text-muted-foreground">
              <GitCommitHorizontal :size="10" class="shrink-0" />
              <span class="font-mono">{{ dep.imageTag }}</span>
              <span v-if="dep.timestamp">&middot; {{ formatRelativeTime(dep.timestamp) }}</span>
              <span v-if="dep.sourceUrl">&middot; via GitHub</span>
            </div>
          </div>

          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button
                variant="ghost"
                size="sm"
                class="h-8 w-8 shrink-0 p-0"
              >
                <MoreVertical :size="16" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                :disabled="deploy.isDeploying"
                @click="handleRollback(dep.imageTag)"
              >
                <RotateCcw :size="14" class="mr-2" />
                Rollback
              </DropdownMenuItem>
              <DropdownMenuItem
                :disabled="deploy.isDeploying"
                @click="handleRedeploy(dep.imageTag)"
              >
                <RefreshCw :size="14" class="mr-2" />
                Redeploy
              </DropdownMenuItem>
              <DropdownMenuSeparator v-if="dep.sourceUrl" />
              <DropdownMenuItem v-if="dep.sourceUrl" as-child>
                <a
                  :href="dep.sourceUrl"
                  target="_blank"
                  rel="noopener"
                >
                  <ExternalLink :size="14" class="mr-2" />
                  View on GitHub
                </a>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </div>

    <!-- No deployment -->
    <EmptyState
      v-else-if="!hasDeployments && !deploy.isDeploying"
      title="No deployment"
      description="This service hasn't been deployed to this environment yet. Click Deploy to get started."
      pattern="diagonal"
    />
  </div>
</template>
