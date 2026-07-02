<script setup lang="ts">
import { computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { ShieldCheck, TriangleAlert, Loader2 } from '@lucide/vue';
import { graphql } from '@/gql';
import EmptyState from '@/components/EmptyState.vue';

const SecretScanReportsDocument = graphql(`
  query SecretScanReports($id: ServiceID!) {
    service(id: $id) {
      id
      secretScanReports {
        scanner
        commit
        scannedAt
        findings {
          rule
          file
          line
          commit
          secret
          author
        }
      }
    }
  }
`);

const props = defineProps<{
  serviceId: string;
}>();

const { result, loading } = useQuery(
  SecretScanReportsDocument,
  () => ({ id: props.serviceId }),
  { fetchPolicy: 'cache-and-network' },
);

const reports = computed(() => result.value?.service?.secretScanReports ?? []);

const scannerLabels: Record<string, string> = {
  gitleaks: 'Gitleaks',
  trufflehog: 'TruffleHog',
};

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
    <div v-if="loading && reports.length === 0" class="flex items-center gap-2 py-8 text-sm text-muted-foreground">
      <Loader2 :size="14" class="animate-spin" />
      Loading scan reports...
    </div>

    <EmptyState
      v-else-if="reports.length === 0"
      title="No scans yet"
      description="Secret scans run automatically with every deployment. Reports for the latest scanned commit show up here."
      pattern="diagonal"
    />

    <div
      v-for="report in reports"
      :key="report.scanner"
      class="rounded-lg border border-border/60 bg-card"
    >
      <div class="flex items-center gap-3 px-4 py-3">
        <ShieldCheck
          v-if="report.findings.length === 0"
          :size="16"
          class="shrink-0 text-[var(--status-ok)]"
        />
        <TriangleAlert
          v-else
          :size="16"
          class="shrink-0 text-[var(--status-warn)]"
        />

        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-foreground">
            {{ scannerLabels[report.scanner] ?? report.scanner }}
          </p>
          <p class="mt-0.5 text-xs text-muted-foreground">
            scanned {{ formatRelativeTime(report.scannedAt) }} &middot;
            <span class="font-mono">{{ shortCommit(report.commit) }}</span>
          </p>
        </div>

        <span
          class="shrink-0 rounded px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide"
          :style="{
            color: report.findings.length === 0 ? 'var(--status-ok)' : 'var(--status-warn)',
            backgroundColor: `color-mix(in srgb, ${report.findings.length === 0 ? 'var(--status-ok)' : 'var(--status-warn)'} 15%, transparent)`,
          }"
        >
          {{ report.findings.length === 0 ? 'Clean' : `${report.findings.length} finding${report.findings.length !== 1 ? 's' : ''}` }}
        </span>
      </div>

      <div v-if="report.findings.length > 0" class="border-t border-border/40">
        <div
          v-for="(finding, idx) in report.findings"
          :key="idx"
          class="flex items-start gap-3 border-b border-border/30 px-4 py-2.5 last:border-b-0"
        >
          <TriangleAlert :size="13" class="mt-0.5 shrink-0 text-[var(--status-warn)]" />
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-foreground">{{ finding.rule }}</p>
            <p class="mt-0.5 truncate font-mono text-xs text-muted-foreground">
              {{ finding.file }}:{{ finding.line }}
              <template v-if="finding.commit"> &middot; {{ shortCommit(finding.commit) }}</template>
            </p>
            <p class="mt-1 rounded bg-muted/60 px-2 py-1 font-mono text-xs text-foreground">{{ finding.secret }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
