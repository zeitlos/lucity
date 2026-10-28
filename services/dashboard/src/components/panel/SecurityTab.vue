<script setup lang="ts">
import RelativeTime from '@/components/RelativeTime.vue';
import QueryError from '@/components/QueryError.vue';
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { ShieldCheck, TriangleAlert, ExternalLink, ChevronDown, ChevronRight, ScanSearch } from '@lucide/vue';
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

const { result, loading, error, refetch } = useQuery(
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

</script>

<template>
  <QueryError
    v-if="error && !result"
    title="Couldn't load security"
    :error="error"
    @retry="refetch()"
  />

  <div v-else class="space-y-8">
    <section v-if="showVulnerabilities" class="space-y-3">
      <div class="space-y-1">
        <h3 class="text-sm font-medium text-muted-foreground">Vulnerabilities</h3>
        <p class="text-sm text-muted-foreground/80">
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
        size="compact"
        :icon="ScanSearch"
        title="No scan yet"
        description="Vulnerabilities in the image show up here once this service has been deployed and scanned."
      />

      <EmptyState
        v-else-if="vulnerabilityReport.summary.total === 0"
        size="compact"
        tone="ok"
        :icon="ShieldCheck"
        title="No known vulnerabilities"
        description="The latest scan of the deployed image found no known CVEs in its packages."
      />

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
          class="w-full justify-between"
          @click="openReport"
        >
          View all {{ vulnerabilityReport.summary.total }} vulnerabilit{{ vulnerabilityReport.summary.total !== 1 ? 'ies' : 'y' }}
          <ChevronRight class="opacity-60" />
        </Button>
      </template>
    </section>

    <section class="space-y-4">
      <div class="space-y-1 text-sm text-muted-foreground/80">
        <h3 class="text-sm font-medium text-muted-foreground">Leaked secrets</h3>
        <p v-if="report">
          Scanned <RelativeTime :value="report.scannedAt" /> at
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
        size="compact"
        :icon="ScanSearch"
        title="No scans yet"
        description="The report for the latest scanned commit shows up here after the first deployment."
      />

      <template v-else>
        <div
          v-if="verifiedFindings.length > 0"
          class="rounded-lg border border-status-danger/40"
        >
          <div
            class="flex items-center gap-2 border-b border-status-danger/20 bg-status-danger/5 px-4 py-2.5 text-sm font-medium text-status-danger"
          >
            <TriangleAlert :size="14" class="shrink-0" />
            {{ verifiedFindings.length }} leaked credential{{ verifiedFindings.length !== 1 ? 's' : '' }} confirmed live. Rotate now.
          </div>
          <div
            v-for="(finding, idx) in verifiedFindings"
            :key="idx"
            class="flex items-center gap-3 border-b border-border/30 px-4 py-3 last:border-b-0"
          >
            <TriangleAlert :size="14" class="shrink-0 text-status-danger" />
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
              size="xs"
              class="shrink-0"
              as-child
            >
              <a :href="finding.url" target="_blank" rel="noopener">
                <ExternalLink />
                Open on GitHub
              </a>
            </Button>
          </div>
        </div>

        <EmptyState
          v-else
          size="compact"
          tone="ok"
          :icon="ShieldCheck"
          title="No leaked credentials"
          description="The latest scan confirmed no live credentials in this repository or its history."
        />

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
                size="xs"
                class="shrink-0"
                as-child
              >
                <a :href="finding.url" target="_blank" rel="noopener">
                  <ExternalLink />
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
