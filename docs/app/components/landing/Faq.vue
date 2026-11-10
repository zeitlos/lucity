<script setup lang="ts">
const questions = [
  { q: 'What can I run on Lucity?', a: 'Web apps, APIs and background workers in whatever you write them in: Node.js, Python, Go, PHP, Ruby, Java, Rust, Elixir, .NET and more. Lucity builds them straight from your repository, and it runs any public container image too.' },
  { q: 'Is it ready for production?', a: 'Yes. Databases run on two machines with automatic failover and continuous backups, rollouts happen without downtime, and a failed deploy never replaces a version that works.' },
  { q: 'What does it cost?', a: 'Hobby is CHF 5 a month and Pro is CHF 25, and each includes that much usage. Beyond it you pay for the CPU, memory and storage your apps actually use. The first 14 days are free.' },
  { q: 'Where does my data live?', a: 'In the EU. Apps and databases run in Germany, files and backups in France, and the company behind Lucity is Swiss.' },
  { q: 'What happens if I want to leave?', a: 'Click Eject. You download your whole setup in a standard format and run it on your own servers or another cloud, without Lucity. Copying your data over is manual work, and the guide walks you through it.' },
  { q: 'Can I host Lucity myself?', a: 'Yes. The whole platform is open source and installs on your own infrastructure, for free.' },
];

const open = ref(0);

function toggle(index: number) {
  open.value = open.value === index ? -1 : index;
}
</script>

<template>
  <section class="mx-auto flex max-w-content flex-wrap items-start gap-x-18 gap-y-8 px-6" aria-labelledby="faq">
    <div class="min-w-[min(280px,100%)] flex-[0_1_380px]">
      <h2 id="faq" class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">Before you ask</h2>
      <p class="mt-4 text-base text-stone-700 dark:text-stone-400">
        Still wondering about something?
        <a href="mailto:christian@zeitlos.software" class="font-semibold text-stone-900 underline underline-offset-4 dark:text-stone-100">Ask us directly</a>
      </p>
    </div>
    <div class="min-w-0 flex-[1_1_640px] border-b border-stone-300 dark:border-stone-800">
      <div v-for="(item, index) in questions" :key="item.q" class="border-t border-stone-300 dark:border-stone-800">
        <h3 class="text-lg/snug font-semibold">
          <button
            :id="`faq-${index}`"
            type="button"
            class="faq-button flex w-full cursor-pointer items-center justify-between gap-6 py-5 text-left text-stone-900 dark:text-stone-100"
            :aria-expanded="open === index"
            :aria-controls="`faq-panel-${index}`"
            @click="toggle(index)"
          >
            <span>{{ item.q }}</span>
            <span
              class="flex size-8 shrink-0 items-center justify-center rounded-full border transition-transform duration-200"
              :class="open === index
                ? 'rotate-45 border-stone-900 bg-stone-900 text-stone-100 dark:border-stone-100 dark:bg-stone-100 dark:text-stone-900'
                : 'border-stone-300 text-stone-800 dark:border-stone-700 dark:text-stone-200'"
              aria-hidden="true"
            >
              <UIcon name="i-lucide-plus" class="size-4" />
            </span>
          </button>
        </h3>
        <p
          v-if="open === index"
          :id="`faq-panel-${index}`"
          role="region"
          :aria-labelledby="`faq-${index}`"
          class="-mt-1.5 max-w-[700px] pr-14 pb-5.5 text-lg/relaxed text-stone-700 dark:text-stone-400"
        >
          {{ item.a }}
        </p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.faq-button:focus-visible {
  outline: 2px solid oklch(0.55 0.15 300);
  outline-offset: 2px;
}
</style>
