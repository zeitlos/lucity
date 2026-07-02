<script setup lang="ts">
import { computed, onMounted, watch } from 'vue';
import {
  Rocket, Check, AlertCircle, Terminal,
  ExternalLink, RefreshCw,
  MoreVertical, ChevronDown, Clock, CircleSlash, TriangleAlert,
} from '@lucide/vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { useNow } from '@vueuse/core';
import { useDeploy } from '@/composables/useDeploy';
import { useBuildLogsPanel, type LogsPanelKind } from '@/composables/useBuildLogsPanel';
import { DeploymentStatus, ReleaseStatus, ScanStatus } from '@/gql/graphql';
import { activeBuild, type Deployment, type Release } from '@/composables/useEnvironment';
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

const emit = defineEmits<{
  (e: 'refetch'): void;
}>();

const deploy = useDeploy();
const logsPanel = useBuildLogsPanel();

const activeDeployment = computed(() => props.service.activeDeployment ?? null);

const sortedReleases = computed(() =>
  [...(props.service.releases ?? [])].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
  ),
);

const historyReleases = computed(() =>
  sortedReleases.value.filter(r => !r.deployment || r.deployment.id !== activeDeployment.value?.id),
);

const activeRelease = computed(() =>
  sortedReleases.value.find(r => r.deployment && r.deployment.id === activeDeployment.value?.id) ?? null,
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
  emit('refetch');
}

async function handleRedeploy() {
  await deploy.startDeploy(props.service.id, props.service.name);
  emit('refetch');
}

function formatRelativeTime(timestamp?: string | null): string {
  if (!timestamp) return 'queued';
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

function shortCommit(commit?: string | null): string | null {
  if (!commit) return null;
  return commit.slice(0, 7);
}

interface StatusMeta {
  label: string;
  color: string;
}

function deploymentStatusMeta(status: DeploymentStatus): StatusMeta {
  switch (status) {
    case DeploymentStatus.Active: return { label: 'Active', color: 'var(--status-ok)' };
    case DeploymentStatus.Deploying: return { label: 'Deploying', color: 'var(--status-progress)' };
    case DeploymentStatus.Failed: return { label: 'Failed', color: 'var(--status-danger)' };
    default: return { label: 'Superseded', color: 'var(--status-neutral)' };
  }
}

function releaseStatusMeta(status: ReleaseStatus): StatusMeta {
  switch (status) {
    case ReleaseStatus.Live: return { label: 'Live', color: 'var(--status-ok)' };
    case ReleaseStatus.Deploying: return { label: 'Deploying', color: 'var(--status-progress)' };
    case ReleaseStatus.Building: return { label: 'Building', color: 'var(--status-progress)' };
    case ReleaseStatus.Queued: return { label: 'Queued', color: 'var(--status-neutral)' };
    case ReleaseStatus.Failed: return { label: 'Failed', color: 'var(--status-danger)' };
    case ReleaseStatus.Cancelled: return { label: 'Cancelled', color: 'var(--status-neutral)' };
    default: return { label: 'Superseded', color: 'var(--status-neutral)' };
  }
}

const IN_FLIGHT_RELEASE_STATUSES = new Set<ReleaseStatus>([
  ReleaseStatus.Queued,
  ReleaseStatus.Building,
  ReleaseStatus.Deploying,
]);

function isInFlight(release: Release): boolean {
  return IN_FLIGHT_RELEASE_STATUSES.has(release.status);
}

type StepStatus = 'queued' | 'running' | 'succeeded' | 'failed' | 'cancelled' | 'skipped' | 'findings';

interface ReleaseStep {
  key: string;
  label: string;
  status: StepStatus;
  detail?: string;
  startedAt?: string | null;
  finishedAt?: string | null;
  logId?: string;
  logKind?: LogsPanelKind;
}

function rolloutStep(deployment: Deployment | null | undefined, live: boolean): ReleaseStep | null {
  if (!deployment) return null;

  let status: StepStatus;

  switch (deployment.status) {
    case DeploymentStatus.Deploying:
      status = 'running';
      break;
    case DeploymentStatus.Failed:
      status = 'failed';
      break;
    case DeploymentStatus.Active:
      status = live && !isReady.value ? 'running' : 'succeeded';
      break;
    default:
      return null;
  }

  return {
    key: 'rollout',
    label: 'Rollout',
    status,
    detail: live && !isReady.value ? `${replicasReady.value}/${replicasDesired.value} ready` : undefined,
    startedAt: status === 'running' ? deployment.createdAt : null,
  };
}

function releaseSteps(release: Release, live = false): ReleaseStep[] {
  const steps: ReleaseStep[] = [];

  if (release.build) {
    steps.push({
      key: 'build',
      label: 'Build',
      status: release.build.status.toLowerCase() as StepStatus,
      startedAt: release.build.startedAt,
      finishedAt: release.build.finishedAt,
      logId: release.build.id,
      logKind: 'build',
    });
  }

  for (const scan of release.scans ?? []) {
    steps.push({
      key: `scan-${scan.scanner}`,
      label: scannerLabels[scan.scanner] ?? scan.scanner,
      status: scanStepStatus(scan.status),
      detail: scan.findingsCount ? `${scan.findingsCount} potential secret${scan.findingsCount !== 1 ? 's' : ''}` : undefined,
      startedAt: scan.startedAt,
      finishedAt: scan.finishedAt,
      logId: scan.id,
      logKind: 'scan',
    });
  }

  if (release.deploy) {
    steps.push({
      key: 'deploy',
      label: 'Deploy',
      status: release.deploy.status.toLowerCase() as StepStatus,
      startedAt: release.deploy.startedAt,
      finishedAt: release.deploy.finishedAt,
      logId: release.deploy.id,
      logKind: 'deploy',
    });
  }

  const rollout = rolloutStep(release.deployment, live);

  if (rollout) {
    steps.push(rollout);
  }

  return steps;
}

const activeSteps = computed<ReleaseStep[]>(() => {
  if (activeRelease.value) {
    return releaseSteps(activeRelease.value, true);
  }

  const rollout = rolloutStep(activeDeployment.value, true);

  return rollout ? [rollout] : [];
});

const scannerLabels: Record<string, string> = {
  gitleaks: 'Gitleaks',
  trufflehog: 'TruffleHog',
};

function scanStepStatus(status: ScanStatus): StepStatus {
  switch (status) {
    case ScanStatus.Clean: return 'succeeded';
    case ScanStatus.Findings: return 'findings';
    case ScanStatus.Failed: return 'failed';
    case ScanStatus.Running: return 'running';
    default: return 'queued';
  }
}

const stepIcons = {
  succeeded: Check,
  failed: AlertCircle,
  running: Spinner,
  queued: Clock,
  cancelled: CircleSlash,
  skipped: CircleSlash,
  findings: TriangleAlert,
} as const;

function stepColor(status: StepStatus): string {
  switch (status) {
    case 'succeeded': return 'var(--status-ok)';
    case 'failed': return 'var(--status-danger)';
    case 'findings': return 'var(--status-warn)';
    case 'running': return 'var(--status-progress)';
    default: return 'var(--status-neutral)';
  }
}

const now = useNow({ interval: 1000 });

function stepDuration(step: ReleaseStep): string | null {
  if (!step.startedAt) return null;
  const start = new Date(step.startedAt).getTime();
  const end = step.finishedAt
    ? new Date(step.finishedAt).getTime()
    : (step.status === 'running' ? now.value.getTime() : null);
  if (end === null) return null;
  const secs = Math.max(0, Math.floor((end - start) / 1000));
  if (secs < 60) return `${secs}s`;
  return `${Math.floor(secs / 60)}m ${secs % 60}s`;
}

function releaseGitHubURL(release: Release): string | null {
  return release.source?.commit.url ?? release.source?.url ?? null;
}

const providerLabels: Record<string, string> = {
  GITHUB: 'GitHub',
  GITLAB: 'GitLab',
  BITBUCKET: 'Bitbucket',
};

function releaseTitle(release: Release): string {
  const message = release.source?.commit.message;
  if (message) return message;
  const sha = release.source?.commit.sha;
  if (sha) return sha.slice(0, 7);
  return release.deployment?.image ?? 'Release';
}

function releaseMeta(release: Release): string {
  const parts: string[] = [formatRelativeTime(release.createdAt)];
  if (release.source) {
    parts.push(`via ${providerLabels[release.source.provider] ?? release.source.provider}`);
  }
  if (release.trigger.actor) {
    parts.push(`by ${release.trigger.actor}`);
  }
  return parts.join(' · ');
}

const replicasReady = computed(() => props.service.replicas?.ready ?? 0);
const replicasDesired = computed(() => props.service.replicas?.desired ?? 0);
const isReady = computed(() => replicasReady.value > 0 && replicasReady.value === replicasDesired.value);
</script>

<template>
  <div class="space-y-4">
    <!-- Deploy Action (source-based services only) -->
    <div v-if="!isImageBased" class="flex items-center gap-3">
      <Button
        :disabled="deploy.isDeploying"
        @click="handleDeploy"
      >
        <Spinner
          v-if="deploy.isDeploying"
          :size="14"
          class="mr-2 animate-spin"
        />
        <Rocket v-else :size="14" class="mr-2" />
        {{ deploy.isDeploying ? 'Building...' : 'Deploy' }}
      </Button>
    </div>

    <!-- Active Deployment Card -->
    <Collapsible
      v-if="activeDeployment"
      :default-open="!isReady"
      class="rounded-lg border bg-card shadow-sm"
    >
      <div class="flex items-center gap-3 pr-2">
        <CollapsibleTrigger class="group flex min-w-0 flex-1 cursor-pointer items-center gap-3 py-3 pl-4 text-left">
          <span
            class="inline-flex w-[84px] shrink-0 items-center justify-center gap-1 rounded py-0.5 text-[10px] font-semibold uppercase tracking-wide"
            :style="{
              color: deploymentStatusMeta(activeDeployment.status).color,
              backgroundColor: `color-mix(in srgb, ${deploymentStatusMeta(activeDeployment.status).color} 15%, transparent)`,
            }"
          >
            <Spinner
              v-if="activeDeployment.status === DeploymentStatus.Deploying || !isReady"
              :size="10"
              class="shrink-0 animate-spin"
            />
            {{ deploymentStatusMeta(activeDeployment.status).label }}
          </span>

          <div class="min-w-0 flex-1">
            <p
              class="truncate text-sm font-medium text-foreground"
              :title="activeDeployment.commitMessage || activeDeployment.commit || activeDeployment.image"
            >
              {{ activeDeployment.commitMessage || shortCommit(activeDeployment.commit) || activeDeployment.image }}
            </p>
            <div class="mt-0.5 flex items-center gap-1.5 text-xs text-muted-foreground">
              <span v-if="activeDeployment.createdAt">{{ formatRelativeTime(activeDeployment.createdAt) }}</span>
              <span v-if="activeRelease?.source">&middot; via {{ providerLabels[activeRelease.source.provider] ?? activeRelease.source.provider }}</span>
              <span v-if="activeRelease?.trigger.actor">&middot; by {{ activeRelease.trigger.actor }}</span>
            </div>
          </div>

          <ChevronDown
            :size="14"
            class="shrink-0 text-muted-foreground transition-transform group-data-[state=open]:rotate-180"
          />
        </CollapsibleTrigger>

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

      <CollapsibleContent>
        <div class="border-t border-border/40 px-4 py-2">
          <div
            v-for="step in activeSteps"
            :key="step.key"
            class="flex items-center gap-3 py-2"
          >
            <div class="flex w-[84px] shrink-0 flex-col">
              <span class="text-sm font-medium text-foreground">{{ step.label }}</span>
              <span v-if="stepDuration(step)" class="text-xs text-muted-foreground">{{ stepDuration(step) }}</span>
            </div>
            <span
              class="flex w-32 shrink-0 items-center gap-1.5 text-sm capitalize"
              :style="{ color: stepColor(step.status) }"
            >
              <component
                :is="stepIcons[step.status]"
                :size="14"
                :class="step.status === 'running' ? 'animate-spin' : ''"
              />
              {{ step.status }}
            </span>
            <span class="flex-1 truncate text-sm text-muted-foreground">{{ step.detail }}</span>
            <Button
              v-if="step.logId && step.logKind"
              variant="outline"
              size="sm"
              class="h-7 shrink-0 gap-1.5 px-2.5"
              @click="logsPanel.open(step.logId, service.name, step.logKind)"
            >
              <Terminal :size="12" />
              Logs
            </Button>
          </div>
        </div>
      </CollapsibleContent>
    </Collapsible>

    <!-- Release history -->
    <div v-if="historyReleases.length > 0" class="space-y-2">
      <h3 class="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        History
      </h3>

      <div class="space-y-2">
        <Collapsible
          v-for="release in historyReleases"
          :key="release.id"
          class="rounded-lg border border-border/60 bg-muted/30"
          :default-open="isInFlight(release)"
        >
          <div class="flex items-center gap-3 pr-2">
            <CollapsibleTrigger class="group flex min-w-0 flex-1 cursor-pointer items-center gap-3 py-3 pl-4 text-left">
              <span
                class="inline-flex w-[84px] shrink-0 items-center justify-center gap-1 rounded py-0.5 text-[10px] font-semibold uppercase tracking-wide"
                :style="{
                  color: releaseStatusMeta(release.status).color,
                  backgroundColor: `color-mix(in srgb, ${releaseStatusMeta(release.status).color} 15%, transparent)`,
                }"
              >
                <Spinner
                  v-if="isInFlight(release)"
                  :size="10"
                  class="shrink-0 animate-spin"
                />
                {{ releaseStatusMeta(release.status).label }}
              </span>

              <div class="min-w-0 flex-1">
                <p
                  class="truncate text-sm font-medium text-foreground"
                  :title="releaseTitle(release)"
                >
                  {{ releaseTitle(release) }}
                </p>
                <div class="mt-0.5 flex items-center gap-1.5 text-xs text-muted-foreground">
                  <span>{{ releaseMeta(release) }}</span>
                </div>
              </div>

              <ChevronDown
                :size="14"
                class="shrink-0 text-muted-foreground transition-transform group-data-[state=open]:rotate-180"
              />
            </CollapsibleTrigger>

            <DropdownMenu v-if="releaseGitHubURL(release)">
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="sm" class="h-8 w-8 shrink-0 p-0">
                  <MoreVertical :size="16" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem as-child>
                  <a
                    :href="releaseGitHubURL(release)!"
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

          <CollapsibleContent>
            <div class="border-t border-border/40 px-4 py-2">
              <template v-if="releaseSteps(release).length > 0">
                <div
                  v-for="step in releaseSteps(release)"
                  :key="step.key"
                  class="flex items-center gap-3 py-2"
                >
                  <div class="flex w-[84px] shrink-0 flex-col">
                    <span class="text-sm font-medium text-foreground">{{ step.label }}</span>
                    <span v-if="stepDuration(step)" class="text-xs text-muted-foreground">{{ stepDuration(step) }}</span>
                  </div>
                  <span
                    class="flex w-32 shrink-0 items-center gap-1.5 text-sm capitalize"
                    :style="{ color: stepColor(step.status) }"
                  >
                    <component
                      :is="stepIcons[step.status]"
                      :size="14"
                      :class="step.status === 'running' ? 'animate-spin' : ''"
                    />
                    {{ step.status }}
                  </span>
                  <span class="flex-1 truncate text-sm text-muted-foreground">{{ step.detail }}</span>
                  <Button
                    v-if="step.logId && step.logKind"
                    variant="outline"
                    size="sm"
                    class="h-7 shrink-0 gap-1.5 px-2.5"
                    @click="logsPanel.open(step.logId, service.name, step.logKind)"
                  >
                    <Terminal :size="12" />
                    Logs
                  </Button>
                </div>
              </template>
              <p v-else class="py-2 text-sm text-muted-foreground">
                No pipeline details for this release.
              </p>
            </div>
          </CollapsibleContent>
        </Collapsible>
      </div>
    </div>

    <!-- No releases yet -->
    <EmptyState
      v-else-if="!deploy.isDeploying && !activeDeployment"
      title="No deployments yet"
      description="This service hasn't been deployed yet. Click Deploy to get started."
      pattern="diagonal"
    />
  </div>
</template>
