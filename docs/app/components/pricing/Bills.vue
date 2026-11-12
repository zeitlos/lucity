<script setup lang="ts">
const receipts = [
  {
    title: 'A side project on Hobby',
    rows: [
      { text: '0.25 vCPU in use, at €20', amount: '5.00' },
      { text: '256 MB in use, at €10 per GB', amount: '2.56' },
      { text: '1 GB of disk, at €0.10', amount: '0.10' },
      { text: '0.5 GB of egress, at €0.02', amount: '0.01' },
    ],
    usage: '7.67',
    plan: 'Hobby',
    fee: '5.00',
    credit: '−5.00',
    total: '7.67',
    verdict: 'Usage went past the credit, so the bill is the usage.',
  },
  {
    title: 'A small app on Pro',
    rows: [
      { text: '0.15 vCPU in use, at €20', amount: '3.00' },
      { text: '1.25 GB in use, at €10', amount: '12.50' },
      { text: '20 GB of disk, at €0.10', amount: '2.00' },
      { text: '10 GB of egress, at €0.02', amount: '0.20' },
    ],
    usage: '17.70',
    plan: 'Pro',
    fee: '25.00',
    credit: '−17.70',
    total: '25.00',
    verdict: 'Usage stayed inside the credit, so the plan is the whole bill.',
  },
];

const notes = [
  { name: 'Inside the credit', text: 'The plan is the whole bill. Nothing is added.' },
  { name: 'Past the credit', text: 'You pay for the usage above it, at the rates on this page.' },
  { name: 'One invoice per workspace', text: 'Every workspace has its own plan and its own invoice.' },
  { name: 'Visible as it adds up', text: 'Billing in the dashboard shows usage so far and the estimated total.' },
];
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="bills">
    <LandingChapterHeading id="bills" beat="keep" promise="Two bills" rest="above the credit and below.">
      You pay the plan or the usage, whichever is higher. These two months show both cases, line by line.
    </LandingChapterHeading>

    <LandingStage tint="keep" class="mt-10 px-4 pt-6 pb-8 sm:mt-14 sm:px-9 sm:pt-12 sm:pb-12">
      <div class="flex flex-wrap items-stretch justify-center gap-x-10 gap-y-6">
        <div v-for="receipt in receipts" :key="receipt.title" class="receipt">
          <p class="receipt-title"><b>{{ receipt.title }}</b><span>One month</span></p>
          <div v-for="row in receipt.rows" :key="row.text" class="row"><span>{{ row.text }}</span><span>{{ row.amount }}</span></div>
          <div class="row sum"><span>Usage</span><span>{{ receipt.usage }}</span></div>
          <div class="row"><span>{{ receipt.plan }} plan</span><span>{{ receipt.fee }}</span></div>
          <div class="row included"><span>Usage included</span><span>{{ receipt.credit }}</span></div>
          <div class="row total"><span>Total</span><span>€{{ receipt.total }}</span></div>
          <p class="verdict">{{ receipt.verdict }}</p>
        </div>
      </div>
    </LandingStage>

    <div class="mt-10 grid grid-cols-[repeat(auto-fit,minmax(min(250px,100%),1fr))] gap-x-10 border-b border-stone-300 dark:border-stone-800">
      <div v-for="note in notes" :key="note.name" class="border-t border-stone-300 py-5.5 dark:border-stone-800">
        <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ note.name }}</b>
        <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ note.text }}</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.receipt {
  --ink: oklch(0.25 0.03 55);
  --muted: oklch(0.45 0.04 55);
  --meta: oklch(0.47 0.04 55);
  flex: 1 1 340px;
  box-sizing: border-box;
  max-width: 440px;
  min-width: 0;
  padding: 28px 30px 26px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
  font-family: var(--font-mono);
  font-size: 14px;
  line-height: 1.5;
  color: var(--ink);
}

.receipt-title {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
  padding-bottom: 14px;
  border-bottom: 1.5px solid var(--ink);
  font-family: var(--font-sans);
  font-size: 17px;
}

.receipt-title b {
  font-weight: 600;
}

.receipt-title span {
  font-size: 14.5px;
  color: var(--meta);
  white-space: nowrap;
}

.row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 5px 0;
}

.row > span:first-child {
  color: var(--muted);
}

.row > span:last-child {
  white-space: nowrap;
}

.sum {
  margin-top: 10px;
  padding-top: 12px;
  border-top: 1px dashed var(--meta);
}

.sum > span:first-child,
.total > span:first-child {
  color: var(--ink);
}

.included,
.included > span:first-child {
  color: oklch(0.43 0.12 160);
}

.total {
  margin-top: 10px;
  padding-top: 12px;
  border-top: 1.5px solid var(--ink);
  font-size: 17px;
  font-weight: 500;
}

.verdict {
  margin-top: 16px;
  font-family: var(--font-sans);
  font-size: 15px;
  line-height: 1.45;
  color: var(--muted);
}

@media (max-width: 480px) {
  .receipt {
    padding: 22px 18px 20px;
    font-size: 13px;
  }
}
</style>
