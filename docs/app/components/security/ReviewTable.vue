<script setup lang="ts">
interface Check {
  label: string;
  to?: string;
  href?: string;
}

interface Row {
  id?: string;
  area: string;
  answer: string;
  check: Check;
}

const rows: Row[] = [
  {
    area: 'Where data lives',
    answer: 'Apps and databases run in Germany. Buckets and database backups are stored in France. Lucity is a Swiss company.',
    check: { label: 'Privacy policy', to: '/legal/privacy-policy' },
  },
  {
    area: 'What leaves the EU',
    answer: 'Requests reach your apps through a global edge network, at a location near each visitor. Platform logs and metadata can be analyzed with an AI provider in the United States, under standard contractual clauses and without training on them.',
    check: { label: 'Data processing agreement', to: '/legal/data-processing-agreement' },
  },
  {
    area: 'Sub-processors',
    answer: 'Four, each listed with its purpose and location. You get 30 days of notice before we add or replace one.',
    check: { label: 'Data processing agreement', to: '/legal/data-processing-agreement' },
  },
  {
    area: 'Isolation',
    answer: 'Workspaces cannot reach each other. The platform runs on machines of its own. Customer apps run as unprivileged containers on shared machines, not in a virtual machine each.',
    check: { label: 'How isolation works', href: '#walls' },
  },
  {
    area: 'In transit',
    answer: 'HTTPS on every domain, with plain HTTP redirected. Traffic between our machines is encrypted. Public database connections are encrypted all the way to the database.',
    check: { label: 'Domains', to: '/docs/domains' },
  },
  {
    area: 'At rest',
    answer: '[ENCRYPTION AT REST: DISKS, BUCKETS AND BACKUPS, TO CONFIRM]',
    check: { label: '[LINK]' },
  },
  {
    area: 'Access',
    answer: 'Two roles per workspace, admin and member. API tokens carry a role, are shown once and can be revoked at once. Deploys from GitHub Actions need no stored token.',
    check: { label: 'Workspaces', to: '/docs/workspaces' },
  },
  {
    area: 'Not there yet',
    answer: 'Single sign-on, an audit log of what members did, custom roles and permissions per environment. Everyone in a workspace can see and change every environment.',
    check: { label: 'Ask what is planned', href: 'mailto:christian@zeitlos.software?subject=Security%20roadmap' },
  },
  {
    area: 'Backups',
    answer: 'Every database write is archived as it happens and can be restored to any second of the last 30 days. Backups are stored with a different provider than the databases.',
    check: { label: 'PostgreSQL', to: '/docs/postgres#backups' },
  },
  {
    area: 'Your code',
    answer: 'A scan for leaked keys on every deploy, and a vulnerability recheck every two hours, for apps Lucity builds. Prebuilt images are not scanned yet.',
    check: { label: 'Above', href: '#keys' },
  },
  {
    area: 'Breaches and deletion',
    answer: 'We notify you without undue delay when a breach affects your data. Data of a deleted project or account is removed within 30 days.',
    check: { label: 'Data processing agreement', to: '/legal/data-processing-agreement' },
  },
  {
    area: 'Certifications',
    answer: '[NONE OF OUR OWN YET, TO CONFIRM. NAME THE CERTIFICATIONS OF THE DATA CENTERS IF WE WANT TO CITE THEM]',
    check: { label: '[LINK]' },
  },
  {
    id: 'report',
    area: 'Reporting a vulnerability',
    answer: '[SECURITY CONTACT AND DISCLOSURE POLICY]',
    check: { label: '[LINK TO security.txt]' },
  },
];

const linkClass = 'font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100';
</script>

<template>
  <section id="review" class="mx-auto max-w-content scroll-mt-28 px-6" aria-labelledby="review-title">
    <div class="flex flex-wrap items-end justify-between gap-x-12 gap-y-4">
      <div class="min-w-0 flex-[1_1_560px]">
        <h2 id="review-title" class="font-display text-4xl/11 text-balance text-stone-900 md:text-5xl/14 lg:text-6xl/17 dark:text-stone-100">Answers for your security review.</h2>
        <p class="mt-6 max-w-[720px] text-xl leading-relaxed text-pretty text-stone-700 dark:text-stone-300">
          The short version for whoever reviews your vendors. Where we do not have something yet, it says so.
        </p>
      </div>
      <p class="text-[15px] text-stone-600 dark:text-stone-400">Last reviewed [DATE]</p>
    </div>

    <div class="review mt-10 border-b border-stone-300 dark:border-stone-800" role="table" aria-labelledby="review-title">
      <div class="review-head" role="row">
        <span role="columnheader">Area</span>
        <span role="columnheader">Answer</span>
        <span role="columnheader">Where to check</span>
      </div>
      <div v-for="row in rows" :id="row.id" :key="row.area" class="review-row scroll-mt-28 border-t border-stone-300 dark:border-stone-800" role="row">
        <div class="font-semibold text-stone-900 dark:text-stone-100" role="rowheader">{{ row.area }}</div>
        <div class="text-stone-800 dark:text-stone-300" role="cell">{{ row.answer }}</div>
        <div class="review-check" role="cell">
          <NuxtLink v-if="row.check.to" :to="row.check.to" :class="linkClass">{{ row.check.label }}</NuxtLink>
          <a v-else-if="row.check.href" :href="row.check.href" :class="linkClass">{{ row.check.label }}</a>
          <span v-else class="text-stone-600 dark:text-stone-400">{{ row.check.label }}</span>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.review {
  font-size: 16px;
  line-height: 1.5;
}

.review-head,
.review-row {
  display: grid;
  grid-template-columns: minmax(150px, 19%) minmax(0, 1fr) 22%;
}

.review-head > *,
.review-row > * {
  padding: 11px 28px 11px 0;
}

.review-head > :last-child,
.review-row > :last-child {
  padding-right: 0;
}

.review-head {
  font-size: 14.5px;
  font-weight: 600;
  color: oklch(0.47 0.04 55);
}

.review-head > * {
  padding-top: 8px;
  padding-bottom: 12px;
}

.dark .review-head {
  color: oklch(0.70 0.03 65);
}

.review-check {
  font-size: 15.5px;
}

@media (max-width: 767px) {
  .review-head {
    display: none;
  }

  .review-row {
    grid-template-columns: minmax(0, 1fr);
    padding: 18px 0 20px;
  }

  .review-row > * {
    padding: 0;
  }

  .review-row > [role="rowheader"] {
    font-size: 18px;
  }

  .review-row > [role="cell"] {
    margin-top: 6px;
  }

  .review-check {
    margin-top: 10px;
  }
}
</style>
