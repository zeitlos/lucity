<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { Download } from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import ContextRail from '@/components/ContextRail.vue';
import EmptyState from '@/components/EmptyState.vue';
import PageToolbar from '@/components/PageToolbar.vue';
import QueryError from '@/components/QueryError.vue';
import SegmentedControl from '@/components/SegmentedControl.vue';
import InvoiceSheet from '@/components/usage/InvoiceSheet.vue';
import UsageProjectRows from '@/components/usage/UsageProjectRows.vue';
import UsageScopeFilter from '@/components/usage/UsageScopeFilter.vue';
import { useAuth } from '@/composables/useAuth';
import { useCapabilities } from '@/composables/useCapabilities';
import { usePageTitle } from '@/composables/usePageTitle';
import { workspaceNavItems } from '@/lib/contextNav';
import {
  dayDate,
  dayIndex,
  download,
  formatCents,
  formatDay,
  isoDay,
  periodLabel,
  projectColors,
  scoped,
  sum,
  tint,
  usageCsv,
  type UsageScope,
} from '@/lib/usage';

const BillingPeriodsDocument = graphql(`
  query BillingPeriods {
    billingPeriods {
      start
      end
      current
    }
  }
`);

const UsageDocument = graphql(`
  query Usage($periodStart: Time!) {
    usage(periodStart: $periodStart) {
      period {
        start
        end
        current
      }
      measuredUntil
      projects {
        projectId
        name
        deletedAt
        amountCents
        daily
        environments {
          environmentId
          name
          deletedAt
          amountCents
          daily
          resources {
            resourceId
            name
            kind
            amountCents
            daily
          }
        }
      }
      invoice {
        number
        status
        issuedAt
        paidAt
        issuer {
          name
          address
        }
        billedTo {
          name
          address
        }
        lines {
          kind
          plan
          quantity
          unit
          unitAmountCents
          amountCents
        }
        subtotalCents
        creditCents
        totalCents
      }
    }
  }
`);

const route = useRoute();
const router = useRouter();
const { isWorkspaceAdmin, activeWorkspace } = useAuth();
const { agents } = useCapabilities();
const railItems = computed(() => workspaceNavItems(isWorkspaceAdmin.value, agents.value));

usePageTitle(() => ['Usage']);

const { result: periodsResult, loading: periodsLoading, error: periodsError, refetch: refetchPeriods } = useQuery(BillingPeriodsDocument);
const periods = computed(() => periodsResult.value?.billingPeriods ?? []);

const period = computed(() => {
  const key = typeof route.query.period === 'string' ? route.query.period : '';
  return periods.value.find(candidate => isoDay(new Date(candidate.start)) === key) ?? periods.value[0];
});

const periodKey = computed({
  get: () => (period.value ? isoDay(new Date(period.value.start)) : ''),
  set: (key: string) => {
    const index = periods.value.findIndex(candidate => isoDay(new Date(candidate.start)) === key);
    router.replace({ query: { ...route.query, period: index > 0 ? key : undefined } });
  },
});

const periodOptions = computed(() => periods.value.map(candidate => ({
  value: isoDay(new Date(candidate.start)),
  label: periodLabel(candidate.start, candidate.end),
})));

const { result, error, refetch } = useQuery(
  UsageDocument,
  () => ({ periodStart: period.value?.start ?? '' }),
  () => ({ enabled: !!period.value }),
);

const usage = computed(() => (result.value?.usage && result.value.usage.period.start === period.value?.start ? result.value.usage : null));
const projects = computed(() => usage.value?.projects ?? []);
const colors = computed(() => projectColors(projects.value));

const scope = ref<UsageScope | null>(null);
const expanded = ref(new Set<string>());

watch(() => period.value?.start, () => {
  scope.value = null;
});

watch(projects, (list) => {
  expanded.value = new Set(list[0] ? [list[0].projectId] : []);
}, { immediate: true });

watch(scope, (value) => {
  if (!value) return;
  const next = new Set(expanded.value);
  next.add(value.project);
  if (value.environment) next.add(value.environment);
  expanded.value = next;
});

const visible = computed(() => scoped(projects.value, scope.value));
const workspaceCents = computed(() => sum(projects.value.map(project => project.amountCents)));
const scopedCents = computed(() => sum(visible.value.map(project => project.amountCents)));

const scopeName = computed(() => {
  if (!scope.value) return '';
  const project = projects.value.find(candidate => candidate.projectId === scope.value?.project);
  const environment = project?.environments.find(candidate => candidate.environmentId === scope.value?.environment);
  return environment ? `${project?.name} / ${environment.name}` : project?.name ?? '';
});

const hero = computed(() => {
  const soFar = usage.value?.period.current ? ' so far' : '';

  if (scope.value) {
    return {
      usageLabel: `${scopeName.value} usage${soFar}`,
      usage: formatCents(scopedCents.value),
      totalLabel: `Workspace usage${soFar}`,
      total: formatCents(workspaceCents.value),
    };
  }

  return {
    usageLabel: `Usage${soFar}`,
    usage: formatCents(workspaceCents.value),
    totalLabel: `Total${soFar}`,
    total: usage.value?.invoice ? formatCents(usage.value.invoice.totalCents) : '',
  };
});

const days = computed(() => usage.value?.projects[0]?.daily.length ?? 0);
const lastDay = computed(() => (usage.value ? dayIndex(usage.value.period.start, usage.value.measuredUntil) : -1));

const range = computed(() => {
  if (!usage.value) return '';
  const start = formatDay(new Date(usage.value.period.start));
  if (!usage.value.period.current) return periodLabel(usage.value.period.start, usage.value.period.end);
  const until = new Date(usage.value.measuredUntil);
  const time = until.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false });
  return `${start} to ${formatDay(until)}, ${time}`;
});

const strip = computed(() => {
  let items: { key: string; label: string; cents: number; color: string }[];
  const project = visible.value[0];

  if (!scope.value || !project) {
    items = projects.value.map(candidate => ({
      key: candidate.projectId,
      label: candidate.name,
      cents: candidate.amountCents,
      color: tint(colors.value.get(candidate.projectId)!, 45),
    }));
  } else if (!scope.value.environment) {
    items = project.environments.map(environment => ({
      key: environment.environmentId,
      label: environment.name,
      cents: environment.amountCents,
      color: colors.value.get(environment.environmentId)!,
    }));
  } else {
    const color = colors.value.get(project.projectId)!;
    items = (project.environments[0]?.resources ?? []).map((resource, i) => ({
      key: `${resource.kind}:${resource.resourceId}`,
      label: resource.name,
      cents: resource.amountCents,
      color: tint(color, [70, 45, 28, 16][Math.min(i, 3)]!),
    }));
  }

  const total = sum(items.map(item => item.cents));
  return items
    .filter(item => item.cents > 0)
    .map(item => ({ ...item, labeled: total > 0 && item.cents / total >= 0.1 }));
});

const stripLabel = computed(() => `${scope.value ? `${scopeName.value} usage split` : 'Usage split by project'}: ${strip.value.map(item => `${item.label} ${formatCents(item.cents)}`).join(', ')}`);

function downloadCsv() {
  if (!usage.value) return;
  const day = isoDay(dayDate(usage.value.period.start, 0));
  download(`usage-${activeWorkspace.value}-${day}.csv`, usageCsv(visible.value, usage.value.period.start), 'text/csv');
}
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div class="relative flex min-h-0 flex-1 py-3 max-md:py-0">
      <ContextRail :items="railItems" active-key="usage" />

      <div class="relative min-w-0 flex-1">
        <div class="flex h-full flex-col overflow-hidden rounded-lg border bg-card shadow-sm max-md:rounded-none max-md:border-0 max-md:shadow-none">
          <PageToolbar title="Usage">
            <Skeleton v-if="periodsLoading && !periodsResult" class="h-7 w-72" />
            <SegmentedControl
              v-else-if="periodOptions.length > 1"
              v-model="periodKey"
              label="Billing period"
              :options="periodOptions"
            />
          </PageToolbar>

          <div class="flex-1 overflow-y-auto max-md:pb-22">
            <QueryError
              v-if="periodsError && !periodsResult"
              class="p-4"
              title="Couldn't load billing periods"
              :error="periodsError"
              @retry="refetchPeriods()"
            />

            <EmptyState
              v-else-if="!periodsLoading && periods.length === 0"
              title="No usage yet"
              description="Usage shows up here once this workspace has a plan."
            />

            <QueryError
              v-else-if="error && !usage"
              class="p-4"
              title="Couldn't load usage"
              :error="error"
              @retry="refetch()"
            />

            <div v-else class="flex flex-col gap-7 px-6 pt-5 max-sm:px-4">
              <div class="flex flex-wrap items-center gap-2">
                <Skeleton v-if="!usage" class="h-9 w-40" />
                <UsageScopeFilter v-else v-model="scope" :projects="projects" :colors="colors" />
                <Button variant="outline" class="ml-auto" :disabled="!usage" @click="downloadCsv">
                  <Download />
                  Download CSV
                </Button>
              </div>

              <section aria-label="Usage and total" class="flex flex-col gap-4">
                <div class="flex flex-col gap-1">
                  <div class="flex flex-wrap gap-x-6 gap-y-2">
                    <div class="flex flex-col gap-0.5 border-r pr-6">
                      <template v-if="usage">
                        <span class="text-sm text-muted-foreground">{{ hero.usageLabel }}</span>
                        <span class="text-4xl font-semibold tabular-nums">{{ hero.usage }}</span>
                      </template>
                      <template v-else>
                        <Skeleton class="my-0.5 h-4 w-24" />
                        <Skeleton class="my-1 h-9 w-36" />
                      </template>
                    </div>
                    <div v-if="!usage || hero.total" class="flex flex-col gap-0.5">
                      <template v-if="usage">
                        <span class="text-sm text-muted-foreground">{{ hero.totalLabel }}</span>
                        <span class="text-4xl font-semibold tabular-nums">{{ hero.total }}</span>
                      </template>
                      <template v-else>
                        <Skeleton class="my-0.5 h-4 w-24" />
                        <Skeleton class="my-1 h-9 w-36" />
                      </template>
                    </div>
                  </div>
                  <span v-if="usage" class="text-xs text-muted-foreground">{{ range }}</span>
                  <Skeleton v-else class="my-0.5 h-3 w-32" />
                </div>

                <Skeleton v-if="!usage" class="h-9 w-full" />
                <div v-else-if="strip.length" role="img" :aria-label="stripLabel" class="flex h-9 gap-0.75">
                  <div
                    v-for="item in strip"
                    :key="item.key"
                    :title="`${item.label} ${formatCents(item.cents)}`"
                    :class="['flex min-w-1.5 items-center justify-between gap-2 overflow-hidden rounded-sm text-sm whitespace-nowrap', item.labeled ? 'px-3' : '']"
                    :style="{ flex: `${item.cents} 1 0px`, background: item.color }"
                  >
                    <template v-if="item.labeled">
                      <span class="truncate font-medium">{{ item.label }}</span>
                      <span class="tabular-nums">{{ formatCents(item.cents) }}</span>
                    </template>
                  </div>
                </div>
              </section>

              <div class="-mx-6 overflow-x-auto px-6 max-sm:-mx-4 max-sm:px-4">
                <div v-if="!usage" class="flex flex-col">
                  <div v-for="row in 4" :key="row" class="grid grid-cols-[minmax(0,1fr)_minmax(12rem,22rem)_7rem] items-center gap-x-6 border-t py-3">
                    <Skeleton class="h-4 w-32" />
                    <Skeleton class="h-7 w-full" />
                    <Skeleton class="ml-auto h-4 w-16" />
                  </div>
                </div>
                <p v-else-if="visible.length === 0" class="border-t py-6 text-sm text-muted-foreground">
                  Nothing used in this period.
                </p>
                <template v-else>
                  <UsageProjectRows
                    v-model:expanded="expanded"
                    :projects="visible"
                    :colors="colors"
                    :total-cents="workspaceCents"
                    :period-start="usage.period.start"
                    :days="days"
                    :last-day="lastDay"
                  />
                  <div class="grid min-w-160 grid-cols-[minmax(0,1fr)_7rem] gap-x-6 border-t py-3 text-sm font-semibold tabular-nums">
                    <span>{{ hero.usageLabel }}</span>
                    <span class="text-right">{{ hero.usage }}</span>
                  </div>
                </template>
              </div>

              <section
                v-if="usage?.invoice"
                :aria-label="usage.period.current ? 'Invoice preview' : 'Invoice'"
                class="-mx-6 flex flex-col gap-5 border-t bg-background px-6 pt-5 pb-14 max-sm:-mx-4 max-sm:px-4"
              >
                <h2 class="text-sm font-medium text-muted-foreground">
                  {{ usage.period.current ? 'Invoice preview' : 'Invoice' }}
                </h2>
                <InvoiceSheet :invoice="usage.invoice" :period-start="usage.period.start" :period-end="usage.period.end" />
              </section>
              <div v-else class="pb-6" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
