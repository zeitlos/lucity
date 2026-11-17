<script setup lang="ts">
import { computed } from 'vue';
import { Download, ReceiptText } from '@lucide/vue';
import DialogIcon from '@/components/DialogIcon.vue';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { InvoiceLineKind, InvoiceStatus, Plan } from '@/gql/graphql';
import { mobileSheetClass } from '@/lib/dialog';
import { formatDate } from '@/lib/format';
import { formatCents, formatQuantity, formatRate, invoiceLineLabels, periodLabel } from '@/lib/usage';

const props = defineProps<{
  invoice: {
    number?: string | null;
    status: InvoiceStatus;
    issuedAt: string;
    paidAt?: string | null;
    pdfUrl?: string | null;
    lines: {
      kind: InvoiceLineKind;
      plan?: Plan | null;
      proration: boolean;
      periodStart?: string | null;
      periodEnd?: string | null;
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
  isAdmin: boolean;
}>();

const open = defineModel<boolean>('open', { required: true });

const planLabels: Partial<Record<Plan, string>> = {
  [Plan.Hobby]: 'Hobby plan',
  [Plan.Pro]: 'Pro plan',
};

const preview = computed(() => props.invoice.status === InvoiceStatus.Preview);

const lines = computed(() => props.invoice.lines.map((line) => {
  if (line.kind === InvoiceLineKind.Plan) {
    const plan = (line.plan && planLabels[line.plan]) ?? invoiceLineLabels[line.kind];
    const time = line.amountCents < 0 ? 'Unused time on' : 'Remaining time on';
    return {
      label: line.proration ? `${time} ${plan}` : plan,
      period: line.periodStart && line.periodEnd ? periodLabel(line.periodStart, line.periodEnd) : '',
      quantity: '',
      rate: '',
      amount: formatCents(line.amountCents),
    };
  }

  return {
    label: invoiceLineLabels[line.kind],
    period: '',
    quantity: `${formatQuantity(line.quantity)} ${line.unit}`,
    rate: `${formatRate(line.unitAmountCents)} per ${line.unit.replace(/s$/, '')}`,
    amount: formatCents(line.amountCents),
  };
}));

const payment = computed(() => {
  if (props.invoice.status === InvoiceStatus.Paid && props.invoice.paidAt) return `Paid on ${formatDate(props.invoice.paidAt)}.`;
  if (preview.value) return `Charged on ${formatDate(props.invoice.issuedAt)}. Usage is counted up to now.`;
  return `Issued on ${formatDate(props.invoice.issuedAt)}, not paid yet.`;
});

const totalLabel = computed(() => ({
  [InvoiceStatus.Preview]: 'Estimated total',
  [InvoiceStatus.Paid]: 'Amount paid',
  [InvoiceStatus.Open]: 'Amount due',
})[props.invoice.status]);
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent :class="[mobileSheetClass, 'max-h-screen overflow-y-auto before:hidden sm:max-w-3xl']">
      <DialogHeader>
        <DialogIcon :icon="ReceiptText" />
        <DialogTitle>{{ preview ? 'Invoice preview' : 'Invoice' }}</DialogTitle>
        <DialogDescription>
          {{ periodLabel(periodStart, periodEnd) }}. {{ payment }}
        </DialogDescription>
      </DialogHeader>

      <dl v-if="invoice.number" class="flex flex-col gap-0.5 text-sm">
        <dt class="text-xs text-muted-foreground">Number</dt>
        <dd class="font-mono text-xs leading-5">{{ invoice.number }}</dd>
      </dl>

      <div class="-mx-6 min-w-0 border-y">
        <Table class="tabular-nums">
          <TableHeader>
            <TableRow class="hover:bg-transparent">
              <TableHead class="first:pl-6">Description</TableHead>
              <TableHead class="text-right max-sm:hidden">Quantity</TableHead>
              <TableHead class="text-right max-sm:hidden">Unit price</TableHead>
              <TableHead class="text-right last:pr-6">Amount</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="(line, i) in lines" :key="i" class="last:border-b-0 hover:bg-transparent">
              <TableCell class="first:pl-6 sm:whitespace-nowrap">
                {{ line.label }}
                <span v-if="line.period" class="block text-xs text-muted-foreground">{{ line.period }}</span>
                <span v-if="line.quantity" class="block text-xs text-muted-foreground sm:hidden">
                  {{ line.quantity }} at {{ line.rate }}
                </span>
              </TableCell>
              <TableCell class="text-right whitespace-nowrap max-sm:hidden">{{ line.quantity }}</TableCell>
              <TableCell class="text-right whitespace-nowrap text-muted-foreground max-sm:hidden">{{ line.rate }}</TableCell>
              <TableCell class="text-right whitespace-nowrap last:pr-6">{{ line.amount }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>

      <div class="grid w-full max-w-75 grid-cols-[1fr_auto] gap-x-6 gap-y-1.5 justify-self-end text-sm tabular-nums">
        <span class="text-muted-foreground">Subtotal</span>
        <span class="text-right">{{ formatCents(invoice.subtotalCents) }}</span>
        <template v-if="invoice.creditCents > 0">
          <span class="text-muted-foreground">Plan credit</span>
          <span class="text-right">{{ formatCents(-invoice.creditCents) }}</span>
        </template>
        <span class="mt-1.5 border-t pt-2.5 font-semibold">{{ totalLabel }}</span>
        <span class="mt-1.5 border-t pt-2.5 text-right font-semibold">{{ formatCents(invoice.totalCents) }}</span>
      </div>

      <p class="text-xs text-muted-foreground">
        All amounts in EUR.
        <template v-if="isAdmin">
          Payment methods and earlier invoices are in
          <RouterLink :to="{ name: 'workspace', params: { drawer: 'billing' } }" class="font-medium text-foreground hover:text-muted-foreground">
            Billing</RouterLink>.
        </template>
      </p>

      <DialogFooter v-if="invoice.pdfUrl">
        <Button as-child variant="outline">
          <a :href="invoice.pdfUrl" target="_blank" rel="noopener noreferrer">
            <Download />
            Download PDF
          </a>
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
