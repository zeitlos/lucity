<script setup lang="ts">
const questions = [
  { q: 'What happens when the trial ends?', a: 'After 14 days or CHF 5 of usage, whichever comes first, the workspace is suspended until you choose a plan. Services stop and nothing can be deployed, but databases, volumes and buckets keep their contents, and everything comes back the way it was.' },
  { q: 'Do unused credits roll over?', a: 'No. Credit left over at the end of a billing period does not carry over to the next one.' },
  { q: 'Can I see what I am spending?', a: 'Yes. Billing in the dashboard shows the plan, when the billing period ends, what the workspace has used so far and the estimated total.' },
  { q: 'Can I switch plans?', a: 'Yes, at any time under Billing. The new plan applies immediately and the fee is prorated for the rest of the period. Only admins of a workspace can change its plan.' },
  { q: 'What if a payment fails?', a: 'The workspace is suspended until the payment method is updated. Your data stays put, and the workspace resumes on its own as soon as the payment goes through.' },
  { q: 'Which currency do you bill in?', a: '[Prices on this page are in CHF, the billing docs say EUR. Pick one and state it here, with VAT.]' },
];

const open = ref(0);

function toggle(index: number) {
  open.value = open.value === index ? -1 : index;
}
</script>

<template>
  <section class="mx-auto flex max-w-content flex-wrap items-start gap-x-18 gap-y-8 px-6" aria-labelledby="billing-faq">
    <div class="min-w-[min(280px,100%)] flex-[0_1_380px]">
      <h2 id="billing-faq" class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">Questions about billing</h2>
      <p class="mt-4 text-base text-stone-700 dark:text-stone-400">
        Still wondering about something?
        <a href="mailto:christian@zeitlos.software?subject=Billing" class="font-semibold text-stone-900 underline underline-offset-4 dark:text-stone-100">Ask about billing</a>
      </p>
    </div>
    <div class="min-w-0 flex-[1_1_640px] border-b border-stone-300 dark:border-stone-800">
      <div v-for="(item, index) in questions" :key="item.q" class="border-t border-stone-300 dark:border-stone-800">
        <h3 class="text-lg/snug font-semibold">
          <button
            :id="`billing-faq-${index}`"
            type="button"
            class="faq-button flex w-full cursor-pointer items-center justify-between gap-6 py-5 text-left text-stone-900 dark:text-stone-100"
            :aria-expanded="open === index"
            :aria-controls="`billing-faq-panel-${index}`"
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
          :id="`billing-faq-panel-${index}`"
          role="region"
          :aria-labelledby="`billing-faq-${index}`"
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
