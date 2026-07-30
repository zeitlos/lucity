<script setup lang="ts">
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { ShieldCheck, TriangleAlert, ExternalLink, ChevronDown, ChevronRight } from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import EmptyState from '@/components/EmptyState.vue';
import { useVulnerabilitiesPanel } from '@/composables/useVulnerabilitiesPanel';
import { severityStyle } from '@/lib/severity';

const ServiceSecurityDocument = graphql(`
  query ServiceSecurity($id: ServiceID!) {
    service(id: $id) {
      id
      sourceUrl
      vulnerabilityReport {
        summary {
          critical
          high
          medium
          low
          unknown
          total
        }
      }
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
  serviceName: string;
}>();

const { result, loading } = useQuery(
  ServiceSecurityDocument,
  () => ({ id: props.serviceId }),
  { fetchPolicy: 'cache-and-network' },
);

const vulnerabilitiesPanel = useVulnerabilitiesPanel();

const showVulnerabilities = computed(() => !!result.value?.service?.sourceUrl);
const vulnerabilityReport = computed(() => result.value?.service?.vulnerabilityReport ?? null);

const summaryBadges = computed(() => {
  const summary = vulnerabilityReport.value?.summary;

  if (!summary) return [];

  return [
    { key: 'CRITICAL', label: 'Critical', count: summary.critical },
    { key: 'HIGH', label: 'High', count: summary.high },
    { key: 'MEDIUM', label: 'Medium', count: summary.medium },
    { key: 'LOW', label: 'Low', count: summary.low },
    { key: 'UNKNOWN', label: 'Unknown', count: summary.unknown },
  ].filter(badge => badge.count > 0);
});

function openReport() {
  vulnerabilitiesPanel.open(props.serviceId, props.serviceName);
}

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
  <div class="space-y-6">
    <section v-if="showVulnerabilities" class="space-y-3">
      <div class="space-y-1">
        <h3 class="text-sm font-semibold text-foreground">Vulnerabilities</h3>
        <p class="text-sm text-muted-foreground">
          Known CVEs in the deployed image. Its operating-system and dependency packages are
          scanned continuously and matched against public vulnerability databases.
        </p>
      </div>

      <div v-if="loading && !vulnerabilityReport" class="space-y-2">
        <Skeleton class="h-8 w-full" />
        <Skeleton class="h-9 w-full" />
      </div>

      <EmptyState
        v-else-if="!vulnerabilityReport"
        title="No scan yet"
        description="Vulnerabilities in the image show up here once this service has been deployed and scanned."
        pattern="diagonal"
      />

      <div
        v-else-if="vulnerabilityReport.summary.total === 0"
        class="flex flex-col items-center justify-center overflow-hidden rounded-lg border border-border bg-card px-8 py-12 text-center"
      >
        <div
          class="mb-4 rounded-full p-4"
          :style="{ backgroundColor: 'color-mix(in srgb, var(--status-ok) 12%, transparent)' }"
        >
          <ShieldCheck :size="32" class="text-[var(--status-ok)]" />
        </div>
        <h3 class="text-sm font-semibold text-foreground">No known vulnerabilities</h3>
        <p class="mt-1 max-w-sm text-sm text-muted-foreground">
          The latest scan of the deployed image found no known CVEs in its packages.
        </p>
      </div>

      <template v-else>
        <div class="flex flex-wrap gap-2">
          <span
            v-for="badge in summaryBadges"
            :key="badge.key"
            class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
            :style="severityStyle(badge.key)"
          >
            <span class="text-sm font-semibold tabular-nums">{{ badge.count }}</span>
            {{ badge.label }}
          </span>
        </div>

        <Button
          variant="outline"
          size="sm"
          class="w-full justify-between"
          @click="openReport"
        >
          View all {{ vulnerabilityReport.summary.total }} vulnerabilit{{ vulnerabilityReport.summary.total !== 1 ? 'ies' : 'y' }}
          <ChevronRight :size="14" class="opacity-60" />
        </Button>
      </template>
    </section>

    <div v-if="showVulnerabilities" class="border-t border-border" />

    <section class="space-y-4">
      <div class="space-y-1 text-sm text-muted-foreground">
        <h3 class="text-sm font-semibold text-foreground">Leaked secrets</h3>
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
    </section>
  </div>
</template>
