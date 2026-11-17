<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { Download, ReceiptText } from '@lucide/vue';
import { graphql } from '@/gql';
import { InvoiceLineKind, InvoiceStatus } from '@/gql/graphql';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import ContextRail from '@/components/ContextRail.vue';
import EmptyState from '@/components/EmptyState.vue';
import PageToolbar from '@/components/PageToolbar.vue';
import QueryError from '@/components/QueryError.vue';
import SegmentedControl from '@/components/SegmentedControl.vue';
import InvoiceDialog from '@/components/usage/InvoiceDialog.vue';
import UsageInvoiceLines from '@/components/usage/UsageInvoiceLines.vue';
import UsageRows from '@/components/usage/UsageRows.vue';
import UsageScopeFilter from '@/components/usage/UsageScopeFilter.vue';
import { useAuth } from '@/composables/useAuth';
import { useCapabilities } from '@/composables/useCapabilities';
import { usePageTitle } from '@/composables/usePageTitle';
import { workspaceNavItems } from '@/lib/contextNav';
import { formatDate } from '@/lib/format';
import { darkTooltipClass } from '@/lib/tooltip';
import {
  breakdownComplete,
  dayDate,
  dayIndex,
  download,
  foldRow,
  foldRows,
  formatCents,
  formatDay,
  formatMonth,
  formatDayTime,
  formatShare,
  isoDay,
  periodLabel,
  projectColors,
  scopeLevel,
  scoped,
  sum,
  tint,
  usageCsv,
  usageRows,
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
        pdfUrl
        lines {
          kind
          plan
          proration
          periodStart
          periodEnd
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

const periodOptions = computed(() => periods.value.map((candidate, index) => ({
  value: isoDay(new Date(candidate.start)),
  label: candidate.current ? 'Current' : periods.value[index - 1]?.current ? 'Previous' : formatMonth(new Date(candidate.start)),
})));

const { result, error, refetch } = useQuery(
  UsageDocument,
  () => ({ periodStart: period.value?.start ?? '' }),
  () => ({ enabled: !!period.value }),
);

const usage = computed(() => (result.value?.usage && result.value.usage.period.start === period.value?.start ? result.value.usage : null));
const projects = computed(() => usage.value?.projects ?? []);
const colors = computed(() => projectColors(projects.value));

const scope = computed<UsageScope | null>({
  get: () => {
    const key = typeof route.query.scope === 'string' ? route.query.scope : '';

    for (const project of projects.value) {
      if (project.projectId === key) return { project: key };
      if (project.environments.some(environment => environment.environmentId === key)) return { project: project.projectId, environment: key };
    }

    return null;
  },
  set: (value) => {
    router.replace({ query: { ...route.query, scope: value ? value.environment ?? value.project : undefined } });
  },
});

const level = computed(() => scopeLevel(scope.value));
const rows = computed(() => foldRows(usageRows(projects.value, scope.value, colors.value)));
const tail = computed(() => (rows.value.tail.length ? foldRow(rows.value.tail, level.value) : null));

const expanded = ref(new Set<string>());
const invoiceOpen = ref(false);

watch(() => rows.value.main[0]?.key, (key) => {
  expanded.value = new Set(key ? [key] : []);
}, { immediate: true });

const visible = computed(() => scoped(projects.value, scope.value));
const workspaceCents = computed(() => sum(projects.value.map(project => project.amountCents)));
const scopedCents = computed(() => sum(visible.value.map(project => project.amountCents)));
const invoicedLines = computed(() => usage.value?.invoice?.lines.filter(line => line.kind !== InvoiceLineKind.Plan) ?? []);
const invoicedCents = computed(() => sum(invoicedLines.value.map(line => line.amountCents)));
const complete = computed(() => !usage.value?.invoice || breakdownComplete(workspaceCents.value, invoicedCents.value));
const headlineCents = computed(() => (complete.value ? scopedCents.value : invoicedCents.value));

const soFar = computed(() => (usage.value?.period.current ? ' so far' : ''));
const usageLabel = computed(() => `Usage${soFar.value}`);

const range = computed(() => {
  if (!usage.value) return '';
  if (!usage.value.period.current) return periodLabel(usage.value.period.start, usage.value.period.end);
  return `${formatDay(new Date(usage.value.period.start))} to ${formatDayTime(new Date(usage.value.measuredUntil))}`;
});

const payment = computed(() => {
  const invoice = usage.value?.invoice;
  if (!invoice) return '';
  if (invoice.status === InvoiceStatus.Paid && invoice.paidAt) return `Paid on ${formatDate(invoice.paidAt)}`;
  if (invoice.status === InvoiceStatus.Preview) return `Charged on ${formatDate(invoice.issuedAt)}`;
  return 'Not paid yet';
});

const days = computed(() => usage.value?.projects[0]?.daily.length ?? 0);
const lastDay = computed(() => (usage.value ? dayIndex(usage.value.period.start, usage.value.measuredUntil) : -1));

const strip = computed(() => {
  const items = rows.value.main.map(row => ({
    key: row.key,
    label: row.name,
    cents: row.amountCents,
    color: scope.value || !row.color ? row.color : tint(row.color, 45),
    scope: row.scope,
  }));

  if (tail.value) {
    items.push({ key: tail.value.key, label: tail.value.name, cents: tail.value.amountCents, color: tail.value.color, scope: undefined });
  }

  const total = sum(items.map(item => item.cents));

  return items
    .filter(item => item.cents > 0)
    .map(item => ({
      ...item,
      share: formatShare(total > 0 ? item.cents / total : 0),
      labeled: total > 0 && item.cents / total >= 0.1,
    }));
});

function pick(item: { key: string; scope?: UsageScope }) {
  if (item.scope) {
    scope.value = item.scope;
    return;
  }

  expanded.value = new Set(expanded.value).add(item.key);
}

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
            <div class="order-last flex w-full items-center gap-2.5 sm:order-none sm:w-auto">
              <Skeleton v-if="periodsLoading && !periodsResult" class="h-7 w-48 max-sm:flex-1" />
              <SegmentedControl
                v-else-if="periodOptions.length > 1"
                v-model="periodKey"
                label="Billing period"
                :options="periodOptions"
              />
            </div>
            <Button v-if="usage?.invoice" variant="outline" class="max-sm:w-9 max-sm:px-0" @click="invoiceOpen = true">
              <ReceiptText />
              <span class="max-sm:sr-only">{{ usage.period.current ? 'Invoice preview' : 'Invoice' }}</span>
            </Button>
            <Button v-if="complete" variant="outline" class="max-sm:w-9 max-sm:px-0" :disabled="!usage" @click="downloadCsv">
              <Download />
              <span class="max-sm:sr-only">Download CSV</span>
            </Button>
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
            >
              <template v-if="isWorkspaceAdmin" #action>
                <Button as-child>
                  <RouterLink :to="{ name: 'workspace', params: { drawer: 'billing' } }">Choose a plan</RouterLink>
                </Button>
              </template>
            </EmptyState>

            <QueryError
              v-else-if="error && !usage"
              class="p-4"
              title="Couldn't load usage"
              :error="error"
              @retry="refetch()"
            />

            <div v-else class="flex flex-col gap-7 px-6 pt-5 pb-8 max-sm:px-4">
              <Skeleton v-if="!usage" class="h-9 w-40" />
              <UsageScopeFilter v-else-if="complete" v-model="scope" :projects="projects" :colors="colors" />

              <section aria-label="Usage and invoice" class="flex flex-col gap-4">
                <div class="flex flex-wrap gap-x-6 gap-y-4">
                  <div class="flex flex-col gap-0.5">
                    <template v-if="usage">
                      <span class="text-sm text-muted-foreground">{{ usageLabel }}</span>
                      <span class="text-4xl font-semibold tabular-nums">{{ formatCents(headlineCents) }}</span>
                      <span class="mt-0.5 text-xs text-muted-foreground">
                        {{ range }}<template v-if="scope && complete"> · {{ formatCents(workspaceCents) }} in this workspace</template>
                      </span>
                    </template>
                    <template v-else>
                      <Skeleton class="my-0.5 h-4 w-24" />
                      <Skeleton class="my-1 h-9 w-36" />
                      <Skeleton class="my-1 h-3 w-32" />
                    </template>
                  </div>
                  <div v-if="usage?.invoice" class="flex flex-col gap-0.5 border-l pl-6">
                    <span class="text-sm text-muted-foreground">{{ usage.period.current ? 'Estimated invoice' : 'Invoice' }}</span>
                    <span class="text-4xl font-semibold tabular-nums">{{ formatCents(usage.invoice.totalCents) }}</span>
                    <span class="mt-0.5 text-xs text-muted-foreground">{{ payment }}</span>
                  </div>
                  <div v-else-if="!usage" class="flex flex-col gap-0.5 border-l pl-6">
                    <Skeleton class="my-0.5 h-4 w-28" />
                    <Skeleton class="my-1 h-9 w-36" />
                    <Skeleton class="my-1 h-3 w-28" />
                  </div>
                </div>

                <Skeleton v-if="!usage" class="h-9 w-full" />
                <div v-else-if="complete && strip.length" role="group" :aria-label="`Usage split by ${level}`" class="flex h-9 gap-0.75">
                  <Tooltip v-for="item in strip" :key="item.key">
                    <TooltipTrigger as-child>
                      <component
                        :is="level === 'resource' ? 'div' : 'button'"
                        :type="level === 'resource' ? undefined : 'button'"
                        :aria-label="`${item.label} ${formatCents(item.cents)}`"
                        :class="[
                          'flex min-w-1.5 items-center justify-between gap-2 overflow-hidden rounded-sm text-sm whitespace-nowrap focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none',
                          item.labeled ? 'px-3' : '',
                          level === 'resource' ? '' : 'transition hover:brightness-95',
                        ]"
                        :style="{ flex: `${item.cents} 1 0px`, background: item.color }"
                        @click="level !== 'resource' && pick(item)"
                      >
                        <template v-if="item.labeled">
                          <span class="truncate font-medium">{{ item.label }}</span>
                          <span class="tabular-nums">{{ formatCents(item.cents) }}</span>
                        </template>
                      </component>
                    </TooltipTrigger>
                    <TooltipContent :class="darkTooltipClass">
                      {{ item.label }} · {{ formatCents(item.cents) }} · {{ item.share }}
                    </TooltipContent>
                  </Tooltip>
                </div>
              </section>

              <div v-if="!usage" class="-mx-4 flex flex-col sm:-mx-6">
                <div
                  v-for="row in 4"
                  :key="row"
                  class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-6 border-t px-4 py-3 sm:grid-cols-[minmax(0,1fr)_minmax(12rem,22rem)_7rem] sm:px-6"
                >
                  <Skeleton class="h-4 w-32" />
                  <Skeleton class="h-7 w-full max-sm:hidden" />
                  <Skeleton class="ml-auto h-4 w-16" />
                </div>
              </div>
              <UsageInvoiceLines
                v-else-if="!complete"
                :lines="invoicedLines"
                :total-label="usageLabel"
                :total-cents="invoicedCents"
              />
              <p v-else-if="rows.main.length === 0" class="-mx-4 border-t px-4 py-6 text-sm text-muted-foreground sm:-mx-6 sm:px-6">
                Nothing used in this period.
              </p>
              <UsageRows
                v-else
                v-model:expanded="expanded"
                :rows="rows.main"
                :tail="rows.tail"
                :level="level"
                :total-label="usageLabel"
                :period-start="usage.period.start"
                :measured-until="usage.measuredUntil"
                :days="days"
                :last-day="lastDay"
                :running="usage.period.current"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <InvoiceDialog
      v-if="usage?.invoice"
      v-model:open="invoiceOpen"
      :invoice="usage.invoice"
      :period-start="usage.period.start"
      :period-end="usage.period.end"
      :is-admin="isWorkspaceAdmin"
    />
  </div>
</template>
