<script setup lang="ts">
import { computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { X, ShieldCheck, ExternalLink, Layers, Package } from '@lucide/vue';
import { onKeyStroke } from '@vueuse/core';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
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
  serviceName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

onKeyStroke('Escape', () => emit('close'));

const { result, loading } = useQuery(
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
  <div class="flex h-full flex-col rounded-lg border bg-card shadow-2xl">
    <div class="flex shrink-0 items-center justify-between border-b px-4 py-3">
      <div class="min-w-0">
        <h2 class="truncate text-sm font-semibold text-foreground">Vulnerabilities</h2>
        <p class="truncate text-xs text-muted-foreground">{{ serviceName }}</p>
      </div>
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        @click="emit('close')"
      >
        <X :size="16" />
      </Button>
    </div>

    <ScrollArea class="flex-1">
      <div class="space-y-4 px-4 py-4">
        <div v-if="loading && !report" class="space-y-2">
          <Skeleton class="h-12 w-full" />
          <Skeleton class="h-12 w-full" />
          <Skeleton class="h-12 w-full" />
          <Skeleton class="h-12 w-full" />
        </div>

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

          <div class="rounded-lg border border-border/60">
            <div
              v-for="(vulnerability, idx) in report.vulnerabilities"
              :key="`${vulnerability.id}-${idx}`"
              class="flex items-start gap-3 border-b border-border/30 px-4 py-3 last:border-b-0"
            >
              <span
                class="mt-0.5 shrink-0 rounded px-1.5 py-0.5 text-[11px] font-semibold uppercase tracking-wide"
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
                    class="inline-flex shrink-0 items-center gap-1 rounded border border-border px-1.5 py-0.5 text-[11px] font-medium text-muted-foreground"
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
                      &rarr; <span class="text-[var(--status-ok)]">{{ pkg.fixedVersion }}</span>
                    </template>
                    <template v-else> &middot; no fix available</template>
                  </p>
                  <p v-if="pkg.path" :title="pkg.path" class="truncate font-mono text-[11px] text-muted-foreground/70">
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

        <div
          v-else
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
      </div>
    </ScrollArea>
  </div>
</template>
