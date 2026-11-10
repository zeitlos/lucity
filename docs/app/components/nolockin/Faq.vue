<script setup lang="ts">
defineProps<{
  title: string;
  items: { q: string; a: string }[];
  askLabel: string;
  askHref: string;
}>();

const open = ref(0);

function toggle(index: number) {
  open.value = open.value === index ? -1 : index;
}
</script>

<template>
  <section class="mx-auto flex max-w-content flex-wrap items-start gap-x-18 gap-y-8 px-6" aria-labelledby="faq">
    <div class="min-w-[min(280px,100%)] flex-[0_1_380px]">
      <h2 id="faq" class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">{{ title }}</h2>
      <p class="mt-4 text-base text-stone-700 dark:text-stone-400">
        Still wondering about something?
        <a :href="askHref" class="font-semibold text-stone-900 underline underline-offset-4 dark:text-stone-100">{{ askLabel }}</a>
      </p>
    </div>
    <div class="min-w-0 flex-[1_1_640px] border-b border-stone-300 dark:border-stone-800">
      <div v-for="(item, index) in items" :key="item.q" class="border-t border-stone-300 dark:border-stone-800">
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
  outline: 2px solid oklch(0.58 0.20 0);
  outline-offset: 2px;
}
</style>
