<script setup lang="ts">
const rates = [
  { name: 'CPU', counted: 'What your services, databases and stores actually use, not what they could use', rate: 'CHF 20', unit: 'per vCPU, per month' },
  { name: 'Memory', counted: 'What they actually use, so a service that sits idle costs very little', rate: 'CHF 10', unit: 'per GB, per month' },
  { name: 'Disk', counted: 'Volumes and database storage, at their full size', rate: 'CHF 0.10', unit: 'per GB, per month' },
  { name: 'Object storage', counted: 'The data your buckets hold', rate: 'CHF 0.03', unit: 'per GB, per month' },
  { name: 'Egress', counted: 'Traffic your services send out through their domains', rate: 'CHF 0.02', unit: 'per GB' },
];
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="rates">
    <LandingChapterHeading id="rates" beat="grow" promise="Five rates" rest="and nothing else on the bill.">
      Everything is metered. No fee per seat, per project, per deploy or per build minute.
    </LandingChapterHeading>

    <table class="rates mt-10 w-full sm:mt-12">
      <thead>
        <tr>
          <th scope="col">Resource</th>
          <th scope="col">What is counted</th>
          <th scope="col" class="rate">Rate</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="rate in rates" :key="rate.name">
          <th scope="row">{{ rate.name }}</th>
          <td>{{ rate.counted }}</td>
          <td class="rate"><b>{{ rate.rate }}</b><span>{{ rate.unit }}</span></td>
        </tr>
      </tbody>
    </table>

    <p class="mt-7 max-w-[820px] text-base text-stone-700 dark:text-stone-400">
      Usage is paid from the plan's credit first. Anything beyond it is added to the next invoice, and credit left over at the end of a billing period does not carry over.
    </p>
  </section>
</template>

<style scoped>
.rates {
  --ink: oklch(0.24 0.03 55);
  --muted: oklch(0.40 0.03 55);
  --meta: oklch(0.47 0.04 55);
  --rule: oklch(0.85 0.03 75);
  border-collapse: collapse;
  font-size: 16px;
}

.dark .rates {
  --ink: oklch(0.95 0.02 80);
  --muted: oklch(0.70 0.03 65);
  --meta: oklch(0.70 0.03 65);
  --rule: oklch(0.30 0.03 55);
}

.rates th,
.rates td {
  padding: 16px 18px 16px 0;
  border-top: 1px solid var(--rule);
  text-align: left;
  vertical-align: baseline;
}

.rates thead th {
  padding-top: 8px;
  padding-bottom: 12px;
  border-top: 0;
  font-size: 14.5px;
  font-weight: 600;
  color: var(--meta);
}

.rates tbody th {
  width: 20%;
  font-size: 18px;
  font-weight: 600;
  color: var(--ink);
}

.rates tbody td {
  color: var(--muted);
}

.rates .rate {
  padding-right: 0;
  text-align: right;
  white-space: nowrap;
}

.rate b {
  margin-right: 12px;
  font-family: var(--font-display);
  font-size: 28px;
  font-weight: 400;
  line-height: 1;
  color: var(--ink);
  font-variant-numeric: tabular-nums;
}

.rate span {
  display: inline-block;
  width: 156px;
  text-align: left;
}

@media (max-width: 767px) {
  .rates,
  .rates tbody {
    display: block;
  }

  .rates thead {
    display: none;
  }

  .rates tbody tr {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 6px 16px;
    align-items: baseline;
    padding: 18px 0;
    border-top: 1px solid var(--rule);
  }

  .rates tbody tr:last-child {
    border-bottom: 1px solid var(--rule);
  }

  .rates tbody th,
  .rates tbody td {
    display: block;
    width: auto;
    padding: 0;
    border: 0;
  }

  .rates tbody td:nth-child(2) {
    grid-column: 1 / -1;
    grid-row: 2;
  }

  .rate b {
    margin-right: 0;
  }

  .rate span {
    display: block;
    width: auto;
    margin-top: 2px;
    font-size: 14.5px;
    text-align: right;
  }
}
</style>
