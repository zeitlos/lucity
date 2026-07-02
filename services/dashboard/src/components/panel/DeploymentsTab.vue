<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue';
import {
  Rocket, Loader2, Check, AlertCircle, Terminal,
  ExternalLink, GitCommitHorizontal, RefreshCw,
  MoreVertical, ChevronDown, Container,
} from '@lucide/vue';
import { useDeploy } from '@/composables/useDeploy';
import { useBuildLogsPanel } from '@/composables/useBuildLogsPanel';
import { BuildStatus, DeploymentStatus, ReleaseStatus } from '@/gql/graphql';
import { activeBuild, type Release } from '@/composables/useEnvironment';
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
}

async function handleRedeploy() {
  await deploy.startDeploy(props.service.id, props.service.name);
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

function shortBuildId(id: string): string {
  const name = id.includes('/') ? id.slice(id.lastIndexOf('/') + 1) : id;
  const trimmed = name.startsWith('build-') ? name.slice(6) : name;
  return trimmed.slice(0, 7);
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
    case DeploymentStatus.Deploying: return { label: 'Deploying', color: 'var(--status-warn)' };
    case DeploymentStatus.Failed: return { label: 'Failed', color: 'var(--status-danger)' };
    default: return { label: 'Superseded', color: 'var(--status-neutral)' };
  }
}

function releaseStatusMeta(status: ReleaseStatus): StatusMeta {
  switch (status) {
    case ReleaseStatus.Live: return { label: 'Live', color: 'var(--status-ok)' };
    case ReleaseStatus.Deploying: return { label: 'Deploying', color: 'var(--status-warn)' };
    case ReleaseStatus.Building: return { label: 'Building', color: 'var(--status-warn)' };
    case ReleaseStatus.Queued: return { label: 'Queued', color: 'var(--status-neutral)' };
    case ReleaseStatus.Failed: return { label: 'Failed', color: 'var(--status-danger)' };
    case ReleaseStatus.Cancelled: return { label: 'Cancelled', color: 'var(--status-neutral)' };
    default: return { label: 'Superseded', color: 'var(--status-neutral)' };
  }
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
const showActiveDetails = ref(false);
</script>

<template>
  <div class="space-y-4">
    <!-- Deploy Action (source-based services only) -->
    <div v-if="!isImageBased" class="flex items-center gap-3">
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

      <button
        v-if="deploy.isDeploying && deploy.buildId"
        class="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground"
        @click="showLogs"
      >
        <Terminal :size="13" />
        Show logs
        <span class="font-mono text-[11px] text-muted-foreground/70">{{ shortBuildId(deploy.buildId) }}</span>
      </button>
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
          <span
            class="mt-1 flex shrink-0 items-center gap-1.5 text-[11px] font-medium"
            :style="{ color: deploymentStatusMeta(activeDeployment.status).color }"
          >
            <span class="h-1.5 w-1.5 rounded-full" :style="{ backgroundColor: deploymentStatusMeta(activeDeployment.status).color }" />
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
              <template v-if="shortCommit(activeDeployment.commit)">
                <GitCommitHorizontal :size="10" class="shrink-0" />
                <a
                  v-if="activeRelease?.source?.commit.url"
                  :href="activeRelease.source.commit.url"
                  target="_blank"
                  rel="noopener"
                  class="font-mono hover:text-foreground hover:underline"
                >{{ shortCommit(activeDeployment.commit) }}</a>
                <span v-else class="font-mono">{{ shortCommit(activeDeployment.commit) }}</span>
              </template>
              <span v-if="activeDeployment.createdAt">&middot; {{ formatRelativeTime(activeDeployment.createdAt) }}</span>
              <span v-if="activeRelease?.source">&middot; via {{ providerLabels[activeRelease.source.provider] ?? activeRelease.source.provider }}</span>
              <span v-if="activeRelease?.trigger.actor">&middot; by {{ activeRelease.trigger.actor }}</span>
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
                <DropdownMenuItem
                  v-if="activeRelease?.build"
                  @click="activeRelease?.build && logsPanel.open(activeRelease.build.id, service.name, 'build')"
                >
                  <Terminal :size="14" class="mr-2" />
                  Build logs
                </DropdownMenuItem>
                <DropdownMenuItem
                  v-if="activeRelease?.deploy"
                  @click="activeRelease?.deploy && logsPanel.open(activeRelease.deploy.id, service.name, 'deploy')"
                >
                  <Terminal :size="14" class="mr-2" />
                  Deploy logs
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="activeRelease?.build || activeRelease?.deploy" />
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

    <!-- Release history -->
    <div v-if="historyReleases.length > 0" class="space-y-2">
      <h3 class="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        History
      </h3>

      <div class="space-y-2">
        <div
          v-for="release in historyReleases"
          :key="release.id"
          class="flex items-center gap-3 rounded-lg border border-border/60 bg-muted/30 px-4 py-3"
        >
          <span
            class="inline-flex w-[84px] shrink-0 justify-center rounded py-0.5 text-[10px] font-semibold uppercase tracking-wide"
            :style="{
              color: releaseStatusMeta(release.status).color,
              backgroundColor: `color-mix(in srgb, ${releaseStatusMeta(release.status).color} 15%, transparent)`,
            }"
          >
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
              <template v-if="release.source">
                <GitCommitHorizontal :size="10" class="shrink-0" />
                <a
                  v-if="release.source.commit.url"
                  :href="release.source.commit.url"
                  target="_blank"
                  rel="noopener"
                  class="font-mono hover:text-foreground hover:underline"
                  @click.stop
                >{{ shortCommit(release.source.commit.sha) }}</a>
                <span v-else class="font-mono">{{ shortCommit(release.source.commit.sha) }}</span>
                <span>&middot;</span>
              </template>
              <span>{{ releaseMeta(release) }}</span>
            </div>
          </div>

          <DropdownMenu v-if="release.build || release.deploy">
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" size="sm" class="h-8 w-8 shrink-0 p-0">
                <MoreVertical :size="16" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                v-if="release.build"
                @click="release.build && logsPanel.open(release.build.id, service.name, 'build')"
              >
                <Terminal :size="14" class="mr-2" />
                Build logs
              </DropdownMenuItem>
              <DropdownMenuItem
                v-if="release.deploy"
                @click="release.deploy && logsPanel.open(release.deploy.id, service.name, 'deploy')"
              >
                <Terminal :size="14" class="mr-2" />
                Deploy logs
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
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
