<script setup lang="ts">
import { computed } from 'vue';
import { InvoiceLineKind, InvoiceStatus, Plan } from '@/gql/graphql';
import { formatDate } from '@/lib/format';
import { formatCents, formatQuantity, formatRate, periodLabel } from '@/lib/usage';
import { cn } from '@/lib/utils';

interface InvoiceParty {
  name: string;
  address: string[];
}

const props = defineProps<{
  invoice: {
    number?: string | null;
    status: InvoiceStatus;
    issuedAt: string;
    paidAt?: string | null;
    issuer: InvoiceParty;
    billedTo: InvoiceParty;
    lines: {
      kind: InvoiceLineKind;
      plan?: Plan | null;
      quantity: number;
      unit: string;
      unitAmountCents: number;
      amountCents: number;
    }[];
    subtotalCents: number;
    creditCents: number;
    totalCents: number;
  };
  periodStart: string;
  periodEnd: string;
}>();

const kindLabels: Record<InvoiceLineKind, string> = {
  [InvoiceLineKind.Plan]: 'Plan',
  [InvoiceLineKind.Cpu]: 'CPU',
  [InvoiceLineKind.Memory]: 'Memory',
  [InvoiceLineKind.Volumes]: 'Volumes',
  [InvoiceLineKind.ObjectStorage]: 'Object storage',
  [InvoiceLineKind.Egress]: 'Egress',
};

const planLabels: Partial<Record<Plan, string>> = {
  [Plan.Hobby]: 'Hobby plan',
  [Plan.Pro]: 'Pro plan',
};

const preview = computed(() => props.invoice.status === InvoiceStatus.Preview);
const period = computed(() => periodLabel(props.periodStart, props.periodEnd));

const lines = computed(() => props.invoice.lines.map((line) => {
  const plan = line.kind === InvoiceLineKind.Plan;
  return {
    label: plan && line.plan ? planLabels[line.plan] ?? kindLabels[line.kind] : kindLabels[line.kind],
    detail: !plan && preview.value ? 'So far' : '',
    quantity: plan ? '1' : `${formatQuantity(line.quantity)} ${line.unit}`,
    unit: `${formatRate(line.unitAmountCents)} per ${line.unit.replace(/s$/, '')}`,
    amount: formatCents(line.amountCents),
  };
}));

const payment = computed(() => {
  if (props.invoice.status === InvoiceStatus.Paid && props.invoice.paidAt) return `Paid on ${formatDate(props.invoice.paidAt)}`;
  if (preview.value) return `Charged on ${formatDate(props.invoice.issuedAt)}`;
  return 'Open';
});

const stamp = computed(() => ({
  [InvoiceStatus.Preview]: 'Preview',
  [InvoiceStatus.Paid]: 'Paid',
  [InvoiceStatus.Open]: 'Open',
})[props.invoice.status]);

const totalLabel = computed(() => ({
  [InvoiceStatus.Preview]: 'Estimated total',
  [InvoiceStatus.Paid]: 'Amount paid',
  [InvoiceStatus.Open]: 'Amount due',
})[props.invoice.status]);
</script>

<template>
  <div class="relative w-full max-w-190 self-center">
    <div aria-hidden="true" class="absolute inset-0 translate-x-1.5 translate-y-2 rotate-1 bg-zinc-100 shadow-xs" />
    <article class="relative flex flex-col gap-9 bg-white px-14 pt-12 pb-10 text-sm text-zinc-900 tabular-nums shadow-lg max-sm:px-6 max-sm:pt-8">
      <div
        :class="cn(
          'pointer-events-none absolute bottom-29 left-14 -rotate-8 rounded-sm border-2 px-3.5 py-1 font-bold tracking-wide opacity-80 max-sm:left-6',
          invoice.status === InvoiceStatus.Paid ? 'text-emerald-600' : 'text-zinc-500',
        )"
      >
        {{ stamp }}
      </div>

      <header class="flex flex-wrap justify-between gap-6">
        <div class="flex flex-col gap-1.5">
          <span v-if="invoice.issuer.name" class="text-xl font-bold tracking-tight">{{ invoice.issuer.name }}</span>
          <span v-if="invoice.issuer.address.length" class="text-zinc-500">
            <template v-for="(line, i) in invoice.issuer.address" :key="i">{{ line }}<br></template>
          </span>
        </div>
        <div class="flex min-w-60 flex-col gap-2.5">
          <span class="text-2xl font-semibold">Invoice</span>
          <dl class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-0.5">
            <dt class="text-zinc-500">Number</dt>
            <dd class="font-mono text-xs leading-5">{{ invoice.number ?? 'Draft' }}</dd>
            <dt class="text-zinc-500">{{ preview ? 'Issue date' : 'Issued' }}</dt>
            <dd>{{ formatDate(invoice.issuedAt) }}</dd>
            <dt class="text-zinc-500">Period</dt>
            <dd>{{ period }}</dd>
          </dl>
        </div>
      </header>

      <div class="flex flex-wrap gap-x-16 gap-y-6">
        <div class="flex flex-col gap-1">
          <span class="text-xs text-zinc-500">Billed to</span>
          <span class="font-semibold">{{ invoice.billedTo.name }}</span>
          <span v-if="invoice.billedTo.address.length" class="text-zinc-500">
            <template v-for="(line, i) in invoice.billedTo.address" :key="i">{{ line }}<br></template>
          </span>
        </div>
        <div class="flex flex-col gap-1">
          <span class="text-xs text-zinc-500">Payment</span>
          <span>{{ payment }}</span>
        </div>
      </div>

      <div class="overflow-x-auto">
        <div role="table" aria-label="Invoice lines" class="min-w-140">
          <div role="row" class="grid grid-cols-[minmax(8rem,1fr)_10rem_13rem_6rem] border-b-2 border-zinc-900 pb-2 text-xs text-zinc-500">
            <span role="columnheader">Description</span>
            <span role="columnheader" class="text-right">Quantity</span>
            <span role="columnheader" class="text-right">Unit price</span>
            <span role="columnheader" class="text-right">Amount</span>
          </div>
          <div
            v-for="(line, i) in lines"
            :key="i"
            role="row"
            class="grid grid-cols-[minmax(8rem,1fr)_10rem_13rem_6rem] border-b border-zinc-200 py-2.5"
          >
            <span role="rowheader" class="flex flex-col">
              <span>{{ line.label }}</span>
              <span v-if="line.detail" class="text-xs text-zinc-500">{{ line.detail }}</span>
            </span>
            <span role="cell" class="text-right whitespace-nowrap">{{ line.quantity }}</span>
            <span role="cell" class="text-right whitespace-nowrap text-zinc-500">{{ line.unit }}</span>
            <span role="cell" class="text-right">{{ line.amount }}</span>
          </div>
        </div>
      </div>

      <div class="grid w-full max-w-75 grid-cols-[1fr_auto] gap-x-6 gap-y-2 self-end">
        <span class="text-zinc-500">Subtotal</span>
        <span class="text-right">{{ formatCents(invoice.subtotalCents) }}</span>
        <span class="text-zinc-500">Plan credit</span>
        <span class="text-right">{{ formatCents(-invoice.creditCents) }}</span>
        <span class="col-span-2 h-0.5 bg-zinc-900" />
        <span class="text-base font-semibold">{{ totalLabel }}</span>
        <span class="text-right text-base font-semibold">{{ formatCents(invoice.totalCents) }}</span>
      </div>

      <footer class="flex justify-between gap-4 border-t border-zinc-200 pt-4 text-xs text-zinc-500">
        <span>All amounts in EUR</span>
        <span>Page 1 of 1</span>
      </footer>
    </article>
  </div>
</template>
