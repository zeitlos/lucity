<script setup lang="ts">
defineProps<{
  id: string;
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
  <section class="mx-auto flex max-w-content flex-wrap items-start gap-x-18 gap-y-8 px-6" :aria-labelledby="id">
    <div class="min-w-[min(280px,100%)] flex-[0_1_380px]">
      <h2 :id="id" class="font-display text-3xl/tight text-balance text-stone-900 lg:text-4xl/tight dark:text-stone-100">{{ title }}</h2>
      <p class="mt-4 text-base text-stone-700 dark:text-stone-400">
        Still wondering about something?
        <a :href="askHref" class="font-semibold text-stone-900 underline decoration-stone-400 underline-offset-4 transition-colors hover:decoration-stone-900 dark:text-stone-100 dark:decoration-stone-600 dark:hover:decoration-stone-100">{{ askLabel }}</a>
      </p>
    </div>
    <div class="min-w-0 flex-[1_1_640px] border-b border-stone-300 dark:border-stone-800">
      <div
        v-for="(item, index) in items"
        :key="item.q"
        class="faq-item border-t border-stone-300 dark:border-stone-800"
        :class="{ 'is-open': open === index }"
      >
        <h3 class="text-lg/snug font-semibold">
          <button
            :id="`${id}-${index}`"
            type="button"
            class="faq-button group flex w-full cursor-pointer items-center justify-between gap-6 py-5 text-left text-stone-900 dark:text-stone-100"
            :aria-expanded="open === index"
            :aria-controls="`${id}-panel-${index}`"
            @click="toggle(index)"
          >
            <span>{{ item.q }}</span>
            <span
              class="faq-icon flex size-8 shrink-0 items-center justify-center rounded-full border"
              :class="open === index
                ? 'rotate-45 border-stone-900 bg-stone-900 text-stone-100 dark:border-stone-100 dark:bg-stone-100 dark:text-stone-900'
                : 'border-stone-300 text-stone-800 group-hover:border-stone-500 dark:border-stone-700 dark:text-stone-200 dark:group-hover:border-stone-500'"
              aria-hidden="true"
            >
              <UIcon name="i-lucide-plus" class="size-4" />
            </span>
          </button>
        </h3>
        <div
          :id="`${id}-panel-${index}`"
          role="region"
          :aria-labelledby="`${id}-${index}`"
          :inert="open !== index"
          class="faq-panel"
        >
          <div class="min-h-0 overflow-hidden">
            <p class="faq-answer max-w-[700px] pb-5.5 sm:pr-14 text-lg/relaxed text-stone-700 dark:text-stone-400">
              {{ item.a }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.faq-button:focus-visible {
  outline: 2px solid oklch(0.55 0.15 300);
  outline-offset: 2px;
}

.faq-icon {
  transition:
    rotate 0.35s cubic-bezier(0.2, 0.8, 0.2, 1),
    background-color 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease;
}

.faq-panel {
  display: grid;
  grid-template-rows: 0fr;
  margin-top: 0;
  transition:
    grid-template-rows 0.35s cubic-bezier(0.2, 0.8, 0.2, 1),
    margin-top 0.35s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.is-open .faq-panel {
  grid-template-rows: 1fr;
  margin-top: -0.375rem;
}

.faq-answer {
  opacity: 0;
  translate: 0 -0.375rem;
  transition:
    opacity 0.2s ease,
    translate 0.35s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.is-open .faq-answer {
  opacity: 1;
  translate: none;
  transition:
    opacity 0.3s ease 0.08s,
    translate 0.35s cubic-bezier(0.2, 0.8, 0.2, 1);
}

@media (prefers-reduced-motion: reduce) {
  .faq-icon,
  .faq-panel,
  .faq-answer,
  .is-open .faq-answer {
    transition: none;
  }
}
</style>
