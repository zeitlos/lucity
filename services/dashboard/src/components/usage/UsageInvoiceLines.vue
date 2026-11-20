<script setup lang="ts">
import { InvoiceLineKind } from '@/gql/graphql';
import { formatCents, formatQuantity, invoiceLineLabels } from '@/lib/usage';

defineProps<{
  lines: {
    kind: InvoiceLineKind;
    quantity: number;
    unit: string;
    amountCents: number;
  }[];
  totalLabel: string;
  totalCents: number;
}>();

const grid = 'grid grid-cols-[minmax(0,1fr)_auto_6rem] items-center gap-x-6 px-4 sm:px-6';
</script>

<template>
  <div class="-mx-4 flex flex-col gap-3 sm:-mx-6">
    <p class="px-4 text-sm text-muted-foreground sm:px-6">No per-project breakdown for this period.</p>

    <div role="table" aria-label="Usage by meter" class="text-sm tabular-nums">
      <div role="row" :class="[grid, 'pb-1.5 text-xs text-muted-foreground']">
        <span role="columnheader" class="font-medium">Meter</span>
        <span role="columnheader" class="text-right font-medium">Quantity</span>
        <span role="columnheader" class="text-right font-medium">Amount</span>
      </div>

      <div v-for="(line, i) in lines" :key="i" role="row" :class="[grid, 'h-11 border-t']">
        <span role="rowheader" class="truncate font-semibold">{{ invoiceLineLabels[line.kind] }}</span>
        <span role="cell" class="text-right whitespace-nowrap text-muted-foreground">{{ formatQuantity(line.quantity) }} {{ line.unit }}</span>
        <span role="cell" class="text-right font-semibold">{{ formatCents(line.amountCents) }}</span>
      </div>

      <div role="row" class="grid grid-cols-[minmax(0,1fr)_auto] gap-x-6 border-t px-4 py-3 font-semibold sm:px-6">
        <span role="rowheader">{{ totalLabel }}</span>
        <span role="cell" class="text-right">{{ formatCents(totalCents) }}</span>
      </div>
    </div>
  </div>
</template>
