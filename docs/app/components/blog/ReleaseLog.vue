<script setup lang="ts">
import { releases, releasesUrl } from './releases';

const props = defineProps<{
  count?: number;
}>();

const shown = computed(() => props.count ? releases.slice(0, props.count) : releases);
</script>

<template>
  <div>
    <div class="timeline mt-7">
      <div v-for="release in shown" :key="release.date" class="timeline-group">
        <div class="timeline-date text-sm font-semibold text-stone-900 dark:text-stone-100">{{ release.date }}</div>
        <div v-for="item in release.items" :key="item.name" class="pb-3.5">
          <b class="block text-lg/6 font-semibold text-stone-900 dark:text-stone-100">{{ item.name }}</b>
          <span class="mt-0.5 block max-w-[620px] text-base text-stone-700 dark:text-stone-400">{{ item.text }}</span>
        </div>
        <a v-if="release.tag" :href="`${releasesUrl}/tag/${release.tag}`" class="inline-flex min-h-8 items-center text-[15px] text-stone-700 underline decoration-1 underline-offset-4 hover:text-stone-900 dark:text-stone-400 dark:hover:text-stone-100">Release notes for {{ release.tag }}</a>
      </div>
    </div>
    <p class="all mt-1.5 text-base">
      <a :href="releasesUrl" class="inline-flex min-h-11 items-center font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">All releases</a>
    </p>
  </div>
</template>

<style scoped>
.timeline {
  --tick: oklch(0.70 0.03 70);
  --tick-major: oklch(0.60 0.03 65);
  --tick-mark: oklch(0.25 0.03 55);
  --inset: 104px;
  --ruler: 56px;
  position: relative;
  padding-left: var(--inset);
}

.dark .timeline {
  --tick: oklch(0.45 0.03 55);
  --tick-major: oklch(0.55 0.03 60);
  --tick-mark: oklch(0.90 0.02 80);
}

.timeline::before {
  content: "";
  position: absolute;
  left: calc(var(--ruler) + 8px);
  top: 0;
  bottom: 0;
  width: 10px;
  border-right: 1px solid var(--tick);
  background: repeating-linear-gradient(to bottom, var(--tick) 0 1px, transparent 1px 8px);
}

.timeline::after {
  content: "";
  position: absolute;
  left: var(--ruler);
  top: 0;
  bottom: 0;
  width: 18px;
  background: repeating-linear-gradient(to bottom, var(--tick-major) 0 1px, transparent 1px 40px);
}

.timeline-group {
  position: relative;
  padding-bottom: 26px;
}

.timeline-group::before {
  content: "";
  position: absolute;
  z-index: 1;
  left: calc(var(--ruler) - var(--inset) - 8px);
  top: 11px;
  width: 26px;
  height: 2px;
  background: var(--tick-mark);
}

.timeline-date {
  position: absolute;
  left: calc(-1 * var(--inset));
  top: 0;
  width: 42px;
  text-align: right;
  line-height: 24px;
  white-space: nowrap;
}

.all {
  margin-left: 104px;
}

@media (max-width: 640px) {
  .timeline {
    --inset: 38px;
    --ruler: 0px;
  }

  .timeline-group::before {
    left: calc(-1 * var(--inset));
    width: 18px;
  }

  .timeline-date {
    position: static;
    display: block;
    width: auto;
    margin-bottom: 6px;
    text-align: left;
    font-size: 15px;
  }

  .all {
    margin-left: 38px;
  }
}
</style>
