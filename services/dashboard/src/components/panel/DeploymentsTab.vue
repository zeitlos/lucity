<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue';
import {
  Rocket, Loader2, Check, AlertCircle, Terminal,
  ExternalLink, GitCommitHorizontal, RefreshCw,
  MoreVertical, ChevronDown, Container, Hammer,
} from 'lucide-vue-next';
import { useDeploy } from '@/composables/useDeploy';
import { useBuildLogsPanel } from '@/composables/useBuildLogsPanel';
import { BuildStatus, DeploymentStatus } from '@/gql/graphql';
import { activeBuild, type Build, type Deployment } from '@/composables/useEnvironment';
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
import type { Service } from '@/composables/useEnvironment';

const props = defineProps<{
  service: Service;
}>();

const deploy = useDeploy();
const logsPanel = useBuildLogsPanel();

function showLogs() {
  if (deploy.buildId) {
    logsPanel.open(deploy.buildId, props.service.name);
  }
}

const activeDeployment = computed(() => props.service.activeDeployment ?? null);
const sortedBuilds = computed(() =>
  [...(props.service.builds ?? [])].sort((a, b) => {
    const at = new Date(a.startedAt).getTime();
    const bt = new Date(b.startedAt).getTime();
    return bt - at;
  }),
);
const sortedDeployments = computed(() =>
  [...(props.service.deployments ?? [])].sort((a, b) => {
    const at = new Date(a.createdAt).getTime();
    const bt = new Date(b.createdAt).getTime();
    return bt - at;
  }),
);

const inFlightBuild = computed(() => activeBuild(props.service));

onMounted(() => {
  const build = inFlightBuild.value;
  if (build && build.id !== deploy.buildId) {
    deploy.pollBuild(build.id);
  }
});

// If the env refetch reveals a new in-flight build, pick it up.
watch(inFlightBuild, (build) => {
  if (build && build.id !== deploy.buildId && !deploy.isDeploying) {
    deploy.pollBuild(build.id);
  }
});

const isImageBased = computed(() => !props.service.sourceUrl);

async function handleDeploy() {
  await deploy.startDeploy(props.service.id, props.service.name);
}

async function handleRedeploy() {
  await deploy.startDeploy(props.service.id, props.service.name);
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

function shortBuildId(id: string): string {
  // build-<hash> → <hash>[:7]
  const trimmed = id.startsWith('build-') ? id.slice(6) : id;
  return trimmed.slice(0, 7);
}

function shortCommit(commit?: string | null): string | null {
  if (!commit) return null;
  return commit.slice(0, 7);
}

function buildDuration(build: Build): string | null {
  if (!build.finishedAt) return null;
  const start = new Date(build.startedAt).getTime();
  const end = new Date(build.finishedAt).getTime();
  const secs = Math.max(0, Math.floor((end - start) / 1000));
  if (secs < 60) return `${secs}s`;
  const mins = Math.floor(secs / 60);
  const remSecs = secs % 60;
  return `${mins}m ${remSecs}s`;
}

function buildStatusVariant(status: BuildStatus): 'default' | 'destructive' | 'secondary' {
  if (status === BuildStatus.Succeeded) return 'default';
  if (status === BuildStatus.Failed) return 'destructive';
  return 'secondary';
}

function deploymentStatusVariant(status: DeploymentStatus): 'default' | 'destructive' | 'secondary' {
  if (status === DeploymentStatus.Active) return 'default';
  if (status === DeploymentStatus.Failed) return 'destructive';
  return 'secondary';
}

function deploymentLabel(deployment: Deployment): string {
  if (deployment.commitMessage) return deployment.commitMessage;
  const short = shortCommit(deployment.commit);
  if (short) return short;
  return deployment.image;
}

const replicasReady = computed(() => props.service.replicas?.ready ?? 0);
const replicasDesired = computed(() => props.service.replicas?.desired ?? 0);
const isReady = computed(() => replicasReady.value > 0 && replicasReady.value === replicasDesired.value);
const showActiveDetails = ref(false);
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
          This service uses a pre-built image. Deployments sync automatically.
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
        {{ deploy.isDeploying ? 'Building...' : 'Deploy' }}
      </Button>

      <Badge
        v-if="deploy.status"
        :variant="buildStatusVariant(deploy.status)"
        :hide-dot="deploy.isDeploying"
      >
        <Loader2
          v-if="deploy.isDeploying"
          :size="12"
          class="mr-1 animate-spin"
        />
        {{ deploy.status }}
      </Badge>
    </div>

    <!-- In-flight build -->
    <div
      v-if="deploy.isDeploying"
      class="rounded-lg border border-border/60 bg-muted/30"
    >
      <div class="flex items-center gap-2.5 px-3 py-2.5">
        <Loader2 :size="14" class="animate-spin text-[var(--primary)]" />
        <div class="min-w-0 flex-1">
          <p class="text-xs font-medium text-foreground">
            {{ deploy.status === BuildStatus.Queued ? 'Queued' : 'Building' }}
          </p>
          <p
            v-if="deploy.buildId"
            class="font-mono text-[11px] text-muted-foreground"
          >
            {{ shortBuildId(deploy.buildId) }}
          </p>
        </div>
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

    <!-- Build failed -->
    <div
      v-if="deploy.status === BuildStatus.Failed"
      class="rounded-lg border border-[var(--status-danger)]/30 bg-[var(--status-danger)]/5 px-3 py-2.5"
    >
      <div class="flex items-start gap-2">
        <AlertCircle
          :size="14"
          class="mt-0.5 shrink-0 text-[var(--status-danger)]"
        />
        <div class="min-w-0 space-y-0.5">
          <p class="text-xs font-medium text-[var(--status-danger)]">Build failed</p>
          <p
            v-if="deploy.error"
            class="break-words font-mono text-[11px] text-muted-foreground"
          >
            {{ deploy.error }}
          </p>
          <button
            v-if="deploy.buildId"
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
            :variant="activeDeployment.status === DeploymentStatus.Active ? 'default' : 'secondary'"
            class="mt-0.5 shrink-0"
          >
            {{ activeDeployment.status }}
          </Badge>

          <div class="min-w-0 flex-1">
            <p
              class="truncate text-sm font-medium text-foreground"
              :title="activeDeployment.commitMessage || activeDeployment.commit || activeDeployment.image"
            >
              {{ activeDeployment.commitMessage || shortCommit(activeDeployment.commit) || activeDeployment.image }}
            </p>
            <div class="mt-0.5 flex items-center gap-1.5 text-xs text-muted-foreground">
              <GitCommitHorizontal v-if="shortCommit(activeDeployment.commit)" :size="10" class="shrink-0" />
              <span v-if="shortCommit(activeDeployment.commit)" class="font-mono">{{ shortCommit(activeDeployment.commit) }}</span>
              <span v-if="activeDeployment.createdAt">&middot; {{ formatRelativeTime(activeDeployment.createdAt) }}</span>
            </div>
          </div>

          <div class="flex shrink-0 items-center gap-1.5">
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
                <DropdownMenuItem :disabled="deploy.isDeploying" @click="handleRedeploy">
                  <RefreshCw :size="14" class="mr-2" />
                  Redeploy
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="service.sourceUrl" />
                <DropdownMenuItem v-if="service.sourceUrl" as-child>
                  <a
                    :href="service.sourceUrl"
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

        <!-- Deployment status expandable -->
        <Collapsible v-model:open="showActiveDetails">
          <CollapsibleTrigger class="flex w-full cursor-pointer items-center gap-2 border-t border-border/40 px-4 py-2.5 text-left">
            <template v-if="isReady">
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
                <Container :size="12" class="shrink-0 text-muted-foreground" />
                <span class="font-mono text-xs text-muted-foreground">{{ activeDeployment.image }}</span>
              </div>
              <div class="text-xs text-muted-foreground">
                <template v-if="service.autoscaling">
                  {{ replicasDesired }} replica{{ replicasDesired !== 1 ? 's' : '' }}
                  (autoscaling {{ service.autoscaling.minReplicas }}&ndash;{{ service.autoscaling.maxReplicas }})
                </template>
                <template v-else>
                  {{ replicasDesired }} replica{{ replicasDesired !== 1 ? 's' : '' }}
                </template>
                <template v-if="isReady"> &middot; healthy</template>
                <template v-else> &middot; {{ replicasReady }}/{{ replicasDesired }} ready</template>
              </div>
            </div>
          </CollapsibleContent>
        </Collapsible>
      </div>
    </div>

    <!-- Deployment history -->
    <div v-if="sortedDeployments.length > 0" class="space-y-2">
      <h3 class="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Deployments
      </h3>

      <div class="space-y-2">
        <div
          v-for="dep in sortedDeployments"
          :key="dep.id"
          class="rounded-lg border border-border/60 bg-muted/30 px-4 py-3"
        >
          <Badge
            v-if="dep.status === DeploymentStatus.Active"
            :variant="deploymentStatusVariant(dep.status)"
            class="mb-1.5 text-[0.65rem]"
          >
            {{ dep.status }}
          </Badge>

          <p
            class="truncate text-sm text-foreground"
            :title="dep.commitMessage || dep.commit"
          >
            {{ deploymentLabel(dep) }}
          </p>
          <div class="mt-0.5 flex items-center gap-1.5 text-xs text-muted-foreground">
            <GitCommitHorizontal v-if="shortCommit(dep.commit)" :size="10" class="shrink-0" />
            <span v-if="shortCommit(dep.commit)" class="font-mono">{{ shortCommit(dep.commit) }}</span>
            <span>&middot; {{ formatRelativeTime(dep.createdAt) }}</span>
            <span v-if="dep.deployedBy">&middot; by {{ dep.deployedBy }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Build history -->
    <div v-if="sortedBuilds.length > 0" class="space-y-2">
      <h3 class="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Builds
      </h3>

      <div class="space-y-2">
        <div
          v-for="build in sortedBuilds"
          :key="build.id"
          class="rounded-lg border border-border/60 bg-muted/30 px-4 py-3"
        >
          <Badge
            :variant="buildStatusVariant(build.status)"
            class="mb-1.5 text-[0.65rem]"
          >
            {{ build.status }}
          </Badge>

          <div class="flex items-center gap-3">
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-1.5 text-sm text-foreground">
                <Hammer :size="12" class="shrink-0 text-muted-foreground" />
                <span class="font-mono">{{ shortBuildId(build.id) }}</span>
              </div>
              <div class="mt-0.5 flex items-center gap-1.5 text-xs text-muted-foreground">
                <span>{{ formatRelativeTime(build.startedAt) }}</span>
                <span v-if="buildDuration(build)">&middot; {{ buildDuration(build) }}</span>
              </div>
            </div>

            <Button
              variant="ghost"
              size="sm"
              class="shrink-0 h-7 text-xs text-muted-foreground"
              @click="logsPanel.open(build.id, service.name)"
            >
              <Terminal :size="13" class="mr-1.5" />
              Logs
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- No builds yet -->
    <EmptyState
      v-else-if="!deploy.isDeploying && !activeDeployment"
      title="No builds yet"
      description="This service hasn't been built yet. Click Deploy to get started."
      pattern="diagonal"
    />
  </div>
</template>
