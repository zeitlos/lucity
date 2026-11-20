<script setup lang="ts">
const appUrl = '/app';

interface Plan {
  name: string;
  who: string;
  price: string;
  per: string;
  included: string;
  items: string[];
  button: string;
  focal?: boolean;
}

const plans: Plan[] = [
  {
    name: 'Trial',
    who: 'To look around, on your personal workspace.',
    price: '14 days',
    per: 'free',
    included: 'Up to €5 of usage, no credit card',
    items: ['Just you', 'One instance per service', 'Up to 1 vCPU and 2 GB per instance', 'PostgreSQL with up to 10 GB of storage'],
    button: 'Start free',
  },
  {
    name: 'Hobby',
    who: 'For side projects and experiments.',
    price: '€5',
    per: 'a month',
    included: '€5 of usage included',
    items: ['Just you', '1 custom domain', 'Up to 3 instances per service', 'Up to 2 vCPU and 4 GB per instance', 'Community support'],
    button: 'Start with Hobby',
  },
  {
    name: 'Pro',
    who: 'For teams shipping to production.',
    price: '€25',
    per: 'a month',
    included: '€25 of usage included',
    items: ['Unlimited members', 'Unlimited custom domains', 'Up to 10 instances per service', 'Up to 4 vCPU and 16 GB per instance', 'FADP/nDSG compliance ready', 'Priority support'],
    button: 'Start with Pro',
    focal: true,
  },
];

const columns = ['Trial', 'Hobby', 'Pro', 'Self-hosted'];

const limits = [
  { name: 'Members', values: ['Just you', 'Just you', 'Unlimited', 'Unlimited'] },
  { name: 'Custom domains', values: ['None', '1', 'Unlimited', 'Unlimited'] },
  { name: 'CPU and memory per instance', values: ['1 vCPU, 2 GB', '2 vCPU, 4 GB', '4 vCPU, 16 GB', '4 vCPU, 16 GB'] },
  { name: 'Instances per service', values: ['1', '3', '10', '10'] },
  { name: 'Volumes', values: ['None', 'Up to 32 GB each', 'Up to 128 GB each', 'Up to 128 GB each'] },
  { name: 'Database storage', values: ['Up to 10 GB', 'Up to 32 GB', 'Up to 128 GB', 'Up to 128 GB'] },
  { name: 'Key-value store', values: ['Up to 1 GB', 'Up to 32 GB', 'Up to 128 GB', 'Up to 128 GB'] },
  { name: 'Support', values: ['Community', 'Community', 'Priority', 'Community'] },
];
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="plans">
    <LandingChapterHeading id="plans" beat="joy" promise="Two plans" rest="and the fee is your credit.">
      Stay inside the credit and the plan is the whole bill. Go beyond it and you pay for the usage above it, nothing more.
    </LandingChapterHeading>

    <div class="mt-10 grid gap-x-12 sm:mt-14 md:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="plan in plans"
        :key="plan.name"
        class="flex flex-col pt-7.5 pb-8.5"
        :class="plan.focal
          ? 'focal mt-4 rounded-[14px] bg-[oklch(0.99_0.006_80)] px-6 sm:px-8 md:col-span-2 lg:col-span-1 lg:mt-0 dark:bg-stone-900'
          : 'border-t border-stone-300 dark:border-stone-800'"
      >
        <h3 class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">{{ plan.name }}</h3>
        <p class="mt-2 text-base text-stone-700 dark:text-stone-400">{{ plan.who }}</p>
        <p class="mt-6.5 font-display text-5xl/none whitespace-nowrap text-stone-900 sm:text-[52px]/none lg:max-xl:text-[44px]/none dark:text-stone-100">
          {{ plan.price }}<small class="font-sans text-[17px] text-stone-600 dark:text-stone-400">&nbsp;{{ plan.per }}</small>
        </p>
        <p class="mt-3 text-base font-semibold text-primary-800 dark:text-primary-400">{{ plan.included }}</p>
        <ul class="ticks mt-5 flex flex-col gap-1.75 text-base text-stone-800 dark:text-stone-300">
          <li v-for="item in plan.items" :key="item" class="flex items-baseline gap-2.5">{{ item }}</li>
        </ul>
        <p class="mt-auto pt-7">
          <a
            :href="`${appUrl}/login`"
            class="inline-flex min-h-15 items-center justify-center gap-3 rounded-lg px-7 text-lg font-semibold transition-colors max-sm:w-full"
            :class="plan.focal
              ? 'border border-transparent bg-primary text-white shadow-sm hover:bg-primary-600'
              : 'border border-stone-300 bg-white/50 text-neutral-800 hover:bg-white dark:border-stone-700 dark:bg-white/5 dark:text-neutral-100 dark:hover:bg-white/10'"
          >
            {{ plan.button }}
            <UIcon v-if="plan.focal" name="i-lucide-arrow-right" class="size-5 shrink-0" />
          </a>
        </p>
      </div>
    </div>

    <table class="limits mt-12 w-full sm:mt-14">
      <caption class="sr-only">What each plan allows</caption>
      <thead>
        <tr>
          <td />
          <th v-for="(column, index) in columns" :key="column" scope="col" :class="{ pro: index === 2 }">{{ column }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="limit in limits" :key="limit.name">
          <th scope="row">{{ limit.name }}</th>
          <td v-for="(value, index) in limit.values" :key="index" :class="{ pro: index === 2 }" :data-plan="columns[index]">{{ value }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style scoped>
.focal {
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
}

.dark .focal {
  box-shadow: 0 0 0 1px oklch(1 0 0 / .12);
}

.ticks li::before {
  content: "";
  flex: none;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: oklch(0.43 0.12 160);
  transform: translateY(-2px);
}

.dark .ticks li::before {
  background: oklch(0.75 0.18 160);
}

.limits {
  --ink: oklch(0.24 0.03 55);
  --muted: oklch(0.40 0.03 55);
  --rule: oklch(0.85 0.03 75);
  --lit: oklch(0.99 0.006 80);
  border-collapse: collapse;
  font-size: 16px;
}

.dark .limits {
  --ink: oklch(0.95 0.02 80);
  --muted: oklch(0.70 0.03 65);
  --rule: oklch(0.30 0.03 55);
  --lit: oklch(0.24 0.03 55);
}

.limits th,
.limits td {
  padding: 16px 18px 16px 0;
  border-top: 1px solid var(--rule);
  text-align: left;
  vertical-align: top;
}

.limits thead th,
.limits thead td {
  padding-top: 8px;
  padding-bottom: 12px;
  border-top: 0;
}

.limits thead th {
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 400;
  color: var(--ink);
}

.limits tbody th {
  width: 26%;
  font-weight: 400;
  color: var(--muted);
}

.limits tbody td {
  font-weight: 600;
  color: var(--ink);
}

.limits .pro {
  padding-left: 18px;
  background: var(--lit);
}

.limits .pro + td,
.limits .pro + th {
  padding-left: 18px;
}

@media (max-width: 899px) {
  .limits,
  .limits tbody {
    display: block;
  }

  .limits thead {
    display: none;
  }

  .limits tbody tr {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 0 12px;
    padding: 18px 0 20px;
    border-top: 1px solid var(--rule);
  }

  .limits tbody tr:last-child {
    border-bottom: 1px solid var(--rule);
  }

  .limits tbody th,
  .limits tbody td {
    display: block;
    width: auto;
    padding: 0;
    border: 0;
  }

  .limits tbody th {
    grid-column: 1 / -1;
    margin-bottom: 12px;
    font-size: 17px;
    font-weight: 600;
    color: var(--ink);
  }

  .limits tbody td {
    font-size: 15px;
    line-height: 1.35;
  }

  .limits tbody td::before {
    content: attr(data-plan);
    display: block;
    margin-bottom: 2px;
    font-size: 13.5px;
    font-weight: 400;
    color: var(--muted);
  }

  .limits .pro,
  .limits .pro + td {
    padding-left: 0;
    background: none;
  }
}

@media (max-width: 599px) {
  .limits tbody tr {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .limits tbody th {
    margin-bottom: 0;
  }
}
</style>
