<script setup lang="ts">
import QueryError from '@/components/QueryError.vue';
import EmptyState from '@/components/EmptyState.vue';
import PanelShell from '@/components/PanelShell.vue';
import { computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { ShieldCheck, ExternalLink, Layers, Package } from '@lucide/vue';
import { graphql } from '@/gql';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Skeleton } from '@/components/ui/skeleton';
import { severityStyle, severityLabel } from '@/lib/severity';

const ServiceVulnerabilitiesDocument = graphql(`
  query ServiceVulnerabilities($id: ServiceID!) {
    service(id: $id) {
      id
      vulnerabilityReport {
        image
        summary {
          critical
          high
          medium
          low
          unknown
          total
        }
        vulnerabilities {
          id
          severity
          source
          title
          reference
          packages {
            name
            installedVersion
            fixedVersion
            path
          }
        }
      }
    }
  }
`);

const props = defineProps<{
  serviceId: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();


const { result, loading, error, refetch } = useQuery(
  ServiceVulnerabilitiesDocument,
  () => ({ id: props.serviceId }),
  { fetchPolicy: 'cache-and-network' },
);

const report = computed(() => result.value?.service?.vulnerabilityReport ?? null);

const summaryBadges = computed(() => {
  const summary = report.value?.summary;

  if (!summary) return [];

  return [
    { key: 'CRITICAL', label: 'Critical', count: summary.critical },
    { key: 'HIGH', label: 'High', count: summary.high },
    { key: 'MEDIUM', label: 'Medium', count: summary.medium },
    { key: 'LOW', label: 'Low', count: summary.low },
    { key: 'UNKNOWN', label: 'Unknown', count: summary.unknown },
  ].filter(badge => badge.count > 0);
});

function sourceLabel(source: string): string {
  switch (source) {
    case 'OPERATING_SYSTEM':
      return 'Base image';
    case 'APPLICATION':
      return 'Dependency';
    default:
      return 'Unknown source';
  }
}
</script>

<template>
  <PanelShell title="Vulnerabilities" @close="emit('close')">

    <ScrollArea class="flex-1">
      <div class="space-y-4 px-4 py-4">
        <div v-if="loading && !report" class="space-y-2">
          <Skeleton class="h-12 w-full" />
          <Skeleton class="h-12 w-full" />
          <Skeleton class="h-12 w-full" />
          <Skeleton class="h-12 w-full" />
        </div>

        <QueryError
          v-else-if="error && !report"
          title="Couldn't load vulnerabilities"
          :error="error"
          @retry="refetch()"
        />

        <template v-else-if="report && report.vulnerabilities.length > 0">
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

          <div class="-mx-4 divide-y border-y">
            <div
              v-for="(vulnerability, idx) in report.vulnerabilities"
              :key="`${vulnerability.id}-${idx}`"
              class="flex items-start gap-3 px-4 py-3"
            >
              <span
                class="mt-0.5 w-16 shrink-0 rounded px-1.5 py-0.5 text-center text-xs font-medium"
                :style="severityStyle(vulnerability.severity)"
              >
                {{ severityLabel(vulnerability.severity) }}
              </span>
              <div class="min-w-0 flex-1">
                <div class="flex items-center justify-between gap-2">
                  <a
                    v-if="vulnerability.reference"
                    :href="vulnerability.reference"
                    target="_blank"
                    rel="noopener"
                    class="inline-flex items-center gap-1 font-mono text-sm font-medium text-foreground hover:underline"
                  >
                    {{ vulnerability.id }}
                    <ExternalLink :size="11" class="opacity-50" />
                  </a>
                  <span v-else class="font-mono text-sm font-medium text-foreground">{{ vulnerability.id }}</span>

                  <span
                    class="inline-flex shrink-0 items-center gap-1 rounded border border-border px-1.5 py-0.5 text-xs font-medium text-muted-foreground"
                  >
                    <Layers v-if="vulnerability.source === 'OPERATING_SYSTEM'" :size="11" />
                    <Package v-else :size="11" />
                    {{ sourceLabel(vulnerability.source) }}
                  </span>
                </div>

                <div v-for="pkg in vulnerability.packages" :key="pkg.name" class="mt-0.5">
                  <p class="truncate font-mono text-xs text-muted-foreground">
                    {{ pkg.name }} {{ pkg.installedVersion }}
                    <template v-if="pkg.fixedVersion">
                      &rarr; <span class="text-status-ok">{{ pkg.fixedVersion }}</span>
                    </template>
                    <template v-else> &middot; no fix available</template>
                  </p>
                  <p v-if="pkg.path" :title="pkg.path" class="truncate font-mono text-xs text-muted-foreground/70">
                    {{ pkg.path }}
                  </p>
                </div>

                <p v-if="vulnerability.title" class="mt-1 text-xs text-muted-foreground">
                  {{ vulnerability.title }}
                </p>
              </div>
            </div>
          </div>
        </template>

        <EmptyState
          v-else
          size="compact"
          tone="ok"
          :icon="ShieldCheck"
          title="No known vulnerabilities"
          description="The latest scan of the deployed image found no known CVEs in its packages."
        />
      </div>
    </ScrollArea>
  </PanelShell>
</template>
