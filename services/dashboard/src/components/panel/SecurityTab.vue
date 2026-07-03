<script setup lang="ts">
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { ShieldCheck, TriangleAlert, ExternalLink, ChevronDown } from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import EmptyState from '@/components/EmptyState.vue';

const SecretScanReportDocument = graphql(`
  query SecretScanReport($id: ServiceID!) {
    service(id: $id) {
      id
      secretScanReport {
        commit
        scannedAt
        findings {
          rule
          file
          line
          commit
          secret
          author
          url
          verified
        }
      }
    }
  }
`);

const props = defineProps<{
  serviceId: string;
}>();

const { result, loading } = useQuery(
  SecretScanReportDocument,
  () => ({ id: props.serviceId }),
  { fetchPolicy: 'cache-and-network' },
);

const report = computed(() => result.value?.service?.secretScanReport ?? null);
const verifiedFindings = computed(() => report.value?.findings.filter(f => f.verified) ?? []);
const possibleFindings = computed(() => report.value?.findings.filter(f => !f.verified) ?? []);

const showPossible = ref(false);

function shortCommit(commit: string): string {
  return commit.slice(0, 7);
}

function formatRelativeTime(timestamp: string): string {
  const diffMs = Date.now() - new Date(timestamp).getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 30) return `${diffDays}d ago`;
  return new Date(timestamp).toLocaleDateString();
}
</script>

<template>
  <div class="space-y-4">
    <div class="space-y-1 text-sm text-muted-foreground">
      <p v-if="report">
        Scanned {{ formatRelativeTime(report.scannedAt) }} at
        <code class="rounded bg-muted px-1 py-0.5 font-mono text-xs">{{ shortCommit(report.commit) }}</code>.
        Every deployment scans the repository and its history for leaked credentials.
      </p>
      <p v-else>Every deployment scans the repository and its history for leaked credentials.</p>
      <p>
        Secrets committed to git stay readable in history even after the file is deleted.
        If something shows up here: rotate the credential first, then rewrite history with a
        tool like git-filter-repo.
      </p>
    </div>

    <div v-if="loading && !report" class="space-y-2">
      <Skeleton class="h-16 w-full" />
      <Skeleton class="h-16 w-full" />
      <Skeleton class="h-16 w-full" />
    </div>

    <EmptyState
      v-else-if="!report"
      title="No scans yet"
      description="The report for the latest scanned commit shows up here after the first deployment."
      pattern="diagonal"
    />

    <template v-else>
      <div
        v-if="verifiedFindings.length > 0"
        class="rounded-lg border border-[var(--status-danger)]/40"
      >
        <div
          class="flex items-center gap-2 border-b border-[var(--status-danger)]/20 px-4 py-2.5 text-sm font-medium text-[var(--status-danger)]"
          :style="{ backgroundColor: 'color-mix(in srgb, var(--status-danger) 6%, transparent)' }"
        >
          <TriangleAlert :size="14" class="shrink-0" />
          {{ verifiedFindings.length }} leaked credential{{ verifiedFindings.length !== 1 ? 's' : '' }} confirmed live. Rotate now.
        </div>
        <div
          v-for="(finding, idx) in verifiedFindings"
          :key="idx"
          class="flex items-center gap-3 border-b border-border/30 px-4 py-3 last:border-b-0"
        >
          <TriangleAlert :size="14" class="shrink-0 text-[var(--status-danger)]" />
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-foreground">{{ finding.rule }}</p>
            <p class="mt-0.5 truncate font-mono text-xs text-muted-foreground">
              {{ finding.file }}:{{ finding.line }}
              <template v-if="finding.commit"> &middot; {{ shortCommit(finding.commit) }}</template>
            </p>
            <p class="mt-1 inline-block rounded bg-muted/60 px-2 py-1 font-mono text-xs text-foreground">{{ finding.secret }}</p>
          </div>
          <Button
            v-if="finding.url"
            variant="outline"
            size="sm"
            class="h-7 shrink-0 gap-1.5 px-2.5"
            as-child
          >
            <a :href="finding.url" target="_blank" rel="noopener">
              <ExternalLink :size="12" />
              Open on GitHub
            </a>
          </Button>
        </div>
      </div>

      <div
        v-else
        class="flex flex-col items-center justify-center overflow-hidden rounded-lg border border-border bg-card px-8 py-[4.5rem] text-center"
      >
        <div
          class="mb-4 rounded-full p-4"
          :style="{ backgroundColor: 'color-mix(in srgb, var(--status-ok) 12%, transparent)' }"
        >
          <ShieldCheck :size="32" class="text-[var(--status-ok)]" />
        </div>
        <h3 class="text-sm font-semibold text-foreground">No leaked credentials</h3>
        <p class="mt-1 max-w-sm text-sm text-muted-foreground">
          The latest scan confirmed no live credentials in this repository or its history.
        </p>
      </div>

      <div v-if="possibleFindings.length > 0">
        <button
          class="flex w-full items-center gap-1.5 py-1 text-sm text-muted-foreground hover:text-foreground"
          @click="showPossible = !showPossible"
        >
          <ChevronDown
            :size="14"
            class="shrink-0 transition-transform"
            :class="showPossible ? 'rotate-180' : ''"
          />
          {{ possibleFindings.length }} possible finding{{ possibleFindings.length !== 1 ? 's' : '' }} that could not be verified
        </button>

        <div v-if="showPossible" class="mt-2 rounded-lg border border-border/60">
          <div
            v-for="(finding, idx) in possibleFindings"
            :key="idx"
            class="flex items-center gap-3 border-b border-border/30 px-4 py-3 last:border-b-0"
          >
            <TriangleAlert :size="14" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">{{ finding.rule }}</p>
              <p class="mt-0.5 truncate font-mono text-xs text-muted-foreground">
                {{ finding.file }}:{{ finding.line }}
                <template v-if="finding.commit"> &middot; {{ shortCommit(finding.commit) }}</template>
              </p>
              <p class="mt-1 inline-block rounded bg-muted/60 px-2 py-1 font-mono text-xs text-foreground">{{ finding.secret }}</p>
            </div>
            <Button
              v-if="finding.url"
              variant="outline"
              size="sm"
              class="h-7 shrink-0 gap-1.5 px-2.5"
              as-child
            >
              <a :href="finding.url" target="_blank" rel="noopener">
                <ExternalLink :size="12" />
                Open on GitHub
              </a>
            </Button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
