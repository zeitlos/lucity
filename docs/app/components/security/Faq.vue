<script setup lang="ts">
const questions = [
  { q: 'Does a finding block my deploy?', a: 'No. The scan for leaked keys runs next to the build and reports what it found. The release ships either way, and a failed build never replaces the version that is already running.' },
  { q: 'Do you use our keys to check them?', a: "Only to ask the key's own provider whether it is valid, with a read-only request. Found keys are stored masked, and only members of your workspace see the findings." },
  { q: 'I rotated the key. Why is the finding still there?', a: 'A finding belongs to a commit, so it stays while that commit is in your history. The mark that a key still works is set when the key is found and is not checked again on later deploys yet. Rotating the key is what makes it harmless.' },
  { q: 'Which apps are checked?', a: 'Every app Lucity builds from your repository, in every environment. Prebuilt images from public registries are not scanned yet.' },
  { q: 'What leaves the EU?', a: 'Requests reach your apps through a global edge network, and some platform logs and metadata are analyzed with an AI provider in the United States under standard contractual clauses. The data processing agreement lists every sub-processor.' },
  { q: 'Can our security team read the code?', a: 'Yes. Lucity is open source, so they can check how scanning and isolation work instead of taking our word for it.' },
];

const open = ref(0);

function toggle(index: number) {
  open.value = open.value === index ? -1 : index;
}
</script>

<template>
  <section class="mx-auto flex max-w-content flex-wrap items-start gap-x-18 gap-y-8 px-6" aria-labelledby="security-faq">
    <div class="min-w-[min(280px,100%)] flex-[0_1_380px]">
      <h2 id="security-faq" class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">What reviewers ask us</h2>
      <p class="mt-4 text-base text-stone-700 dark:text-stone-400">
        Still wondering about something?
        <a href="mailto:christian@zeitlos.software?subject=Security%20question" class="font-semibold text-stone-900 underline underline-offset-4 dark:text-stone-100">Ask us</a>
      </p>
    </div>
    <div class="min-w-0 flex-[1_1_640px] border-b border-stone-300 dark:border-stone-800">
      <div v-for="(item, index) in questions" :key="item.q" class="border-t border-stone-300 dark:border-stone-800">
        <h3 class="text-lg/snug font-semibold">
          <button
            :id="`security-faq-${index}`"
            type="button"
            class="faq-button flex w-full cursor-pointer items-center justify-between gap-6 py-5 text-left text-stone-900 dark:text-stone-100"
            :aria-expanded="open === index"
            :aria-controls="`security-faq-panel-${index}`"
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
          :id="`security-faq-panel-${index}`"
          role="region"
          :aria-labelledby="`security-faq-${index}`"
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
