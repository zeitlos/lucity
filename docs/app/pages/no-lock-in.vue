<script setup lang="ts">
definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

useSeo({
  title: 'No lock-in',
  description: 'Every Lucity project has an Eject button. It downloads your services, databases and domains in a standard format any major cloud can run, without Lucity. Here is exactly what happens if you press it.',
  type: 'website',
});

const checks = [
  'No account, no license key, no call home',
  'Your project keeps running on Lucity',
  'Open source, top to bottom',
];

const promises = [
  { name: 'A copy, whenever you like', text: 'Ejecting changes nothing on Lucity. Take an export to keep one, to try your own cluster first, or to move for good.' },
  { name: 'Runs where Kubernetes runs', text: 'A cluster of your own, a managed one from a big cloud, or another provider. The list of what it needs is further down.' },
  { name: 'Honest about the rest', text: 'The export is your configuration, not your data. Moving for good is manual work, and this page says which.' },
];

const ledgers = [
  {
    title: 'In the zip',
    items: [
      { name: 'Services', text: 'with their images, variables, instances, resources and private hostnames' },
      { name: 'Domains', text: 'with their routes and certificate settings' },
      { name: 'Databases, key-value stores and volumes', text: 'configured the same, created empty' },
      { name: 'A build script', text: 'for the images Lucity built from your code' },
    ],
  },
  {
    title: 'You bring',
    items: [
      { name: 'A cluster', text: 'on Kubernetes 1.33 or newer, with nodes that support user namespaces' },
      { name: 'The open source parts it leans on', text: 'depending on what the project uses, listed below' },
      { name: 'A registry', text: 'for the images you rebuild' },
      { name: 'Your data', text: 'copied with pg_dump and any S3 tool. Volumes have no download, so their files come out through your own service' },
      { name: 'A maintenance window', text: 'because there is no live migration. Writes stop while the data is copied' },
    ],
  },
  {
    title: 'Yours from then on',
    items: [
      { name: 'Builds and deploys', text: 'a push no longer deploys anything, so your CI builds and runs helm upgrade' },
      { name: 'Database backups', text: 'backups and point-in-time recovery are set up again on your side' },
      { name: 'Separation between environments', text: 'network rules and resource budgets live outside the chart' },
      { name: 'Logs, metrics and a dashboard', text: 'bring the tools you already like' },
    ],
  },
];

const needs = [
  { has: 'Services', needs: 'Kubernetes 1.33 or newer, on nodes with user namespaces (Linux 6.3+, containerd 2.0+ or CRI-O)' },
  { has: 'Databases, key-value stores or volumes', needs: 'A default StorageClass' },
  { has: 'PostgreSQL databases', needs: 'The CloudNativePG operator' },
  { has: 'Domains', needs: 'A Gateway API implementation and a Gateway reachable from the internet' },
  { has: 'Custom domains', needs: 'ListenerSet support on that Gateway, and cert-manager 1.20+ with an HTTP-01 ClusterIssuer' },
  { has: 'Autoscaling', needs: 'metrics-server' },
];

const days = [
  { day: '1', name: 'Push, and it is live.', text: 'A repository becomes a running app with an address and HTTPS.' },
  { day: '30', name: 'A database in one click.', text: 'Linked to your services, with no passwords copied around.' },
  { day: '365', name: 'Production, with a team.', text: "A customer's domain, real traffic and colleagues with access." },
  { day: '1000', name: 'A reason to change something.', text: 'Compliance wants your own hardware, or you have grown a platform team.' },
];

const exits = [
  { label: 'Stay', name: 'We run it.', text: 'Keep shipping on Lucity. Nothing about your app changes as it grows, and you can take an export any time, just to have one.', line: 'border-solid' },
  { label: 'Eject', name: 'You run your project.', text: 'Install the export on a cluster that fits the list above. Builds, backups and monitoring become yours.', line: 'border-dashed', link: { to: '/docs/eject', label: 'The eject guide' } },
  { label: 'Self-host', name: 'You run the whole platform.', text: 'Lucity is open source. Install it with the same two Helm charts we run, and keep the dashboard, builds and push to deploy. Budget an afternoon for the first install.', line: 'border-dotted', link: { to: '/docs/self-hosting', label: 'Self-hosting' } },
];

const questions = [
  { q: 'Does ejecting change anything on Lucity?', a: 'No. The project keeps running exactly as before, so you can eject to keep a copy, to try your own cluster first, or to move for good. The export is a snapshot: changes you make on Lucity afterwards need a fresh one.' },
  { q: 'Do I still need a Lucity account afterwards?', a: 'No. Nothing in the export calls home, needs a license key or talks to Lucity.' },
  { q: 'Does my data come along?', a: 'No. The export is configuration. Databases, key-value stores and volumes come up empty on your cluster, and bucket files stay in their buckets. The guide shows how to copy each of them.' },
  { q: 'Is there downtime when I move for good?', a: 'Yes, a short one that you plan. There is no live migration: writes have to stop while the data is copied, and custom domains wait a minute or two for new certificates after the DNS switch.' },
  { q: 'Can I run the whole platform myself instead?', a: 'Yes. Lucity is open source, so you can install the full platform on your own cluster and keep the dashboard, builds and push to deploy. It is not a one-command install, and the self-hosting guide says what it takes.' },
];

const readNext = [
  { label: 'The full eject guide', to: '/docs/eject' },
  { label: 'Self-hosting the platform', to: '/docs/self-hosting' },
  { label: 'The software deployment trilemma', to: '/blog/the-software-deployment-trilemma' },
  { label: 'The source on GitHub', to: 'https://github.com/zeitlos/lucity' },
];

const linkClass = 'font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100';
</script>

<template>
  <div class="overflow-x-clip">
    <section class="mx-auto max-w-content px-6 pt-12 sm:pt-28" aria-labelledby="top">
      <h1 id="top" class="load-in font-display text-4xl/[1.15] text-stone-900 sm:text-5xl/[1.1] md:text-6xl/[1.1] lg:text-7xl/[1.1] dark:text-stone-100" style="--load-order: 0">
        The way out is <span class="relative whitespace-nowrap">built in<img src="/img/underline-keep.svg" alt="" class="pointer-events-none absolute bottom-0 left-0 h-[0.32em] w-full"></span>.
      </h1>

      <p class="load-in mt-8 max-w-[820px] text-xl leading-relaxed text-stone-700 sm:text-2xl sm:leading-relaxed dark:text-stone-300" style="--load-order: 1">
        Every Lucity project has an Eject button. It downloads your services, databases and domains in a standard format any major cloud can run, without Lucity. You will probably never press it. Here is exactly what happens if you do.
      </p>

      <div class="load-in mt-8 flex flex-col gap-3 sm:mt-11 sm:flex-row sm:items-center sm:gap-4" style="--load-order: 2">
        <NuxtLink
          to="/docs/eject"
          class="flex items-center justify-center gap-3 rounded-lg border border-transparent bg-primary px-7 py-3.5 text-lg font-semibold text-white shadow-sm transition-colors hover:bg-primary-600 sm:py-4 sm:text-xl"
        >
          Read the eject guide
          <UIcon name="i-lucide-arrow-right" class="size-5.5 shrink-0" />
        </NuxtLink>
        <NuxtLink
          :to="{ hash: '#sample' }"
          class="flex items-center justify-center rounded-lg border border-stone-300 bg-white/50 px-7 py-3.5 text-lg font-semibold text-neutral-800 transition-colors hover:bg-white sm:py-4 sm:text-xl dark:border-stone-700 dark:bg-white/5 dark:text-neutral-100 dark:hover:bg-white/10"
        >
          Download a sample export
        </NuxtLink>
      </div>

      <ul class="load-in mt-5 flex flex-wrap gap-x-7 gap-y-1.5 text-base text-stone-600 sm:mt-6 dark:text-stone-400" style="--load-order: 3">
        <li v-for="check in checks" :key="check" class="flex items-center gap-2">
          <UIcon name="i-lucide-check" class="size-4.5 shrink-0 text-primary-700 dark:text-primary" />
          {{ check }}
        </li>
      </ul>

      <NolockinHeroScene class="load-in mt-12 sm:mt-18" style="--load-order: 4; --load-blur: 0px" />

      <div class="mt-10 grid grid-cols-[repeat(auto-fit,minmax(min(300px,100%),1fr))] gap-x-14 border-b border-stone-300 sm:mt-14 dark:border-stone-800">
        <div v-for="promise in promises" :key="promise.name" class="border-t border-stone-300 py-5.5 dark:border-stone-800">
          <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ promise.name }}</b>
          <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ promise.text }}</span>
        </div>
      </div>
    </section>

    <section class="mx-auto mt-24 max-w-content px-6 sm:mt-38" aria-labelledby="zip">
      <NolockinChapterHeading id="zip" promise="One zip" rest="nothing of ours inside.">
        The export is the chart every Lucity environment is deployed with, plus one file per environment that says how yours is set up. Click through it.
      </NolockinChapterHeading>

      <LandingStage tint="keep" class="mt-14 px-4 pt-5 pb-6 sm:p-10 sm:pb-11">
        <NolockinZipExplorer />
      </LandingStage>

      <p class="mt-7 flex max-w-[900px] items-baseline gap-3 text-base text-stone-700 dark:text-stone-300">
        <UIcon name="i-lucide-triangle-alert" class="size-4.5 shrink-0 translate-y-[3px] text-[oklch(0.48_0.11_70)] dark:text-[oklch(0.80_0.12_75)]" />
        <span>The archive holds your variables in plain text, secrets included, along with the passwords of your key-value stores. Keep it where you would keep a password, and out of public repositories.</span>
      </p>
    </section>

    <section class="mx-auto mt-24 max-w-content px-6 sm:mt-38" aria-labelledby="takes">
      <NolockinChapterHeading id="takes" promise="An honest list" rest="of what moving takes.">
        Ejecting copies your configuration. It is not a migration. This is what the zip carries, what you bring, and what becomes your job once we are out of the picture.
      </NolockinChapterHeading>

      <div class="mt-14 grid grid-cols-[repeat(auto-fit,minmax(min(320px,100%),1fr))] gap-x-14 gap-y-12">
        <div v-for="ledger in ledgers" :key="ledger.title">
          <h3 class="font-display text-2xl/tight text-stone-900 dark:text-stone-100">{{ ledger.title }}</h3>
          <ul class="mt-4 border-b border-stone-300 dark:border-stone-800">
            <li v-for="item in ledger.items" :key="item.name" class="border-t border-stone-300 py-3.5 text-base text-stone-700 dark:border-stone-800 dark:text-stone-400">
              <b class="block text-lg/snug font-semibold text-stone-900 dark:text-stone-100">{{ item.name }}</b>
              {{ item.text }}
            </li>
          </ul>
        </div>
      </div>

      <div class="mt-18 flex flex-wrap items-start gap-x-18 gap-y-6">
        <div class="min-w-0 max-w-[420px] flex-[1_1_320px]">
          <h3 class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">What the cluster needs</h3>
          <p class="mt-3.5 text-lg/normal text-stone-700 dark:text-stone-400">
            Only what your project uses. Every part is open source, and the readme in the download lists them again.
          </p>
          <p class="mt-4 text-base">
            <NuxtLink to="/docs/eject#prerequisites" :class="linkClass">The full list, with versions</NuxtLink>
          </p>
        </div>
        <div class="min-w-0 flex-[2_1_560px]">
          <table class="needs w-full border-collapse text-left text-base">
          <thead>
            <tr>
              <th scope="col">If the project has</th>
              <th scope="col">The cluster needs</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="need in needs" :key="need.has">
              <th scope="row">{{ need.has }}</th>
              <td>{{ need.needs }}</td>
            </tr>
          </tbody>
          </table>
        </div>
      </div>
    </section>

    <section class="mx-auto mt-24 max-w-content px-6 sm:mt-38" aria-labelledby="day">
      <NolockinChapterHeading id="day" promise="Day 1000" rest="and still your call.">
        Most platforms get expensive the day you outgrow them, because your setup lives in their dashboard. Here the next step is a choice you make, not a project you start.
      </NolockinChapterHeading>

      <ol class="days mt-12 sm:mt-16">
        <li v-for="item in days" :key="item.day" class="day">
          <i aria-hidden="true" />
          <span class="day-number"><small>day</small><b>{{ item.day }}</b></span>
          <p class="text-base text-stone-700 dark:text-stone-400">
            <strong class="mb-0.5 block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ item.name }}</strong>
            {{ item.text }}
          </p>
        </li>
      </ol>

      <p class="mt-14 max-w-[900px] font-display text-3xl/tight text-stone-900 sm:mt-18 lg:text-4xl/tight dark:text-stone-100">
        From here, three ways to go on. It is the same chart underneath all of them.
      </p>

      <div class="mt-9 grid grid-cols-[repeat(auto-fit,minmax(min(300px,100%),1fr))] gap-x-14 gap-y-9">
        <div v-for="exit in exits" :key="exit.label" class="border-t-2 border-[oklch(0.58_0.20_0)] pt-6" :class="exit.line">
          <p class="text-[15px] font-semibold text-[oklch(0.48_0.17_0)] dark:text-[oklch(0.76_0.14_0)]">{{ exit.label }}</p>
          <h3 class="mt-1 font-display text-3xl/[1.15] text-stone-900 dark:text-stone-100">{{ exit.name }}</h3>
          <p class="mt-2.5 text-base text-stone-700 dark:text-stone-400">
            {{ exit.text }}
            <NuxtLink v-if="exit.link" :to="exit.link.to" class="font-medium whitespace-nowrap text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">{{ exit.link.label }}</NuxtLink>
          </p>
        </div>
      </div>
    </section>

    <FaqSection
      id="faq"
      class="mt-20 sm:mt-34"
      title="Questions about leaving"
      :items="questions"
      ask-label="Ask us"
      ask-href="mailto:christian@zeitlos.software?subject=Ejecting"
    />

    <section id="sample" class="mx-auto mt-24 max-w-content scroll-mt-28 px-6 pb-20 sm:mt-38 sm:pb-30" aria-labelledby="sample-heading">
      <div class="flex flex-wrap items-center gap-x-18 gap-y-10 border-y border-stone-300 py-10 sm:py-14 dark:border-stone-800">
        <div class="min-w-0 flex-[1_1_520px]">
          <h2 id="sample-heading" class="max-w-[760px] font-display text-4xl/11 text-balance text-stone-900 md:text-5xl/14 lg:text-6xl/17 dark:text-stone-100">
            Read the way out before you come in.
          </h2>
          <p class="mt-5 max-w-[640px] text-xl leading-relaxed text-pretty text-stone-700 dark:text-stone-300">
            Download the export of our demo project, shop. It is the same zip the Eject button gives you, with the chart, the values and the build script. If you do not like what is inside, do not sign up.
          </p>
          <div class="mt-8 flex flex-wrap items-center gap-x-5 gap-y-4">
            <a
              href="#sample"
              class="flex items-center justify-center gap-3 rounded-lg border border-transparent bg-primary px-7 py-3.5 text-lg font-semibold text-white shadow-sm transition-colors hover:bg-primary-600 max-sm:w-full sm:py-4 sm:text-xl"
            >
              Download a sample export
              <UIcon name="i-lucide-download" class="size-5 shrink-0" />
            </a>
            <span class="text-base text-stone-600 dark:text-stone-400">shop-ejected.zip · [SIZE] · [SAMPLE EXPORT LINK]</span>
          </div>
        </div>
        <nav aria-label="Read next" class="min-w-[min(280px,100%)] flex-[0_1_360px] text-lg">
          <p class="mb-1.5 text-base font-semibold text-stone-900 dark:text-stone-100">Read next</p>
          <NuxtLink v-for="item in readNext" :key="item.to" :to="item.to" class="flex py-2" :class="linkClass">{{ item.label }}</NuxtLink>
        </nav>
      </div>
    </section>
  </div>
</template>

<style scoped>
.needs th,
.needs td {
  padding: 16px 18px 16px 0;
  border-top: 1px solid var(--color-stone-300);
  vertical-align: top;
}

.needs thead th {
  padding-top: 8px;
  padding-bottom: 12px;
  border-top: 0;
  font-size: 14.5px;
  font-weight: 600;
  color: var(--color-stone-600);
}

.needs tbody th {
  width: 50%;
  padding-right: 24px;
  font-weight: 600;
  color: var(--color-stone-900);
}

.needs td {
  color: var(--color-stone-700);
}

.dark .needs th,
.dark .needs td {
  border-color: var(--color-stone-800);
}

.dark .needs thead th {
  color: var(--color-stone-400);
}

.dark .needs tbody th {
  color: var(--color-stone-100);
}

.dark .needs td {
  color: var(--color-stone-400);
}

.days {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 28px 0;
}

.day {
  position: relative;
  padding: 26px 32px 8px 0;
  border-top: 2px solid oklch(0.58 0.20 0);
}

.day i {
  position: absolute;
  left: 0;
  top: -7px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: oklch(0.58 0.20 0);
}

.day-number small {
  display: block;
  font-size: 15px;
  line-height: 1.55;
  color: var(--color-stone-600);
}

.day-number b {
  display: block;
  font-family: var(--font-display);
  font-size: 52px;
  font-weight: 400;
  line-height: 1;
  color: var(--color-stone-900);
}

.day p {
  margin-top: 12px;
}

.dark .day-number small {
  color: var(--color-stone-400);
}

.dark .day-number b {
  color: var(--color-stone-100);
}

@media (max-width: 1023px) {
  .days {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 639px) {
  .days {
    grid-template-columns: minmax(0, 1fr);
    gap: 0;
  }

  .day {
    margin-left: 5px;
    padding: 0 0 30px 24px;
    border-top: 0;
    border-left: 2px solid oklch(0.58 0.20 0);
  }

  .day:last-child {
    padding-bottom: 4px;
    border-image: linear-gradient(to bottom, oklch(0.58 0.20 0) 18px, transparent 18px) 1;
  }

  .day i {
    left: -7px;
    top: 13px;
  }

  .day-number {
    display: flex;
    align-items: baseline;
    gap: 8px;
  }

  .day-number b {
    font-size: 40px;
  }

  .day p {
    margin-top: 6px;
  }

  .needs,
  .needs tbody,
  .needs tr,
  .needs th,
  .needs td {
    display: block;
  }

  .needs thead {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
  }

  .needs tbody th {
    width: auto;
    padding: 14px 0 0;
  }

  .needs td {
    padding: 2px 0 14px;
    border-top: 0;
  }
}
</style>
