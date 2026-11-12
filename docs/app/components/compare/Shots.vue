<script setup lang="ts">
import manifest from '~~/public/img/compare/manifest.json';
import { formatComparisonDate } from '~/utils/comparisons';
import type { Capture, Shot } from '~/utils/comparisons/types';

const props = defineProps<{
  platform: string;
  name: string;
  icon: string;
  shots: Shot[];
}>();

const captures = manifest as unknown as Record<string, Record<string, Capture> | undefined>;

const active = ref(0);
const current = computed(() => props.shots[active.value]!);

const sides = computed(() => [
  { key: props.platform, name: props.name, icon: props.icon, lucity: false, capture: captures[props.platform]?.[current.value.name] },
  { key: 'lucity', name: 'Lucity', icon: '', lucity: true, capture: captures.lucity?.[current.value.name] },
]);

function duration(seconds = 0) {
  const minutes = Math.floor(seconds / 60);
  return `${minutes}:${String(Math.round(seconds % 60)).padStart(2, '0')}`;
}

function caption(capture?: Capture) {
  if (!capture) {
    return 'Not captured yet';
  }
  const date = formatComparisonDate(capture.capturedAt);
  return capture.kind === 'video' ? `Recorded ${date}, ${duration(capture.seconds)} in real time` : `Captured ${date}`;
}
</script>

<template>
  <div>
    <div role="tablist" aria-label="Screens to compare" class="mt-10 flex flex-wrap gap-x-7 border-b border-default">
      <button
        v-for="(shot, index) in shots"
        :id="`shot-tab-${shot.name}`"
        :key="shot.name"
        type="button"
        role="tab"
        :aria-selected="index === active"
        :aria-controls="`shot-panel-${shot.name}`"
        class="-mb-px min-h-12 border-b-2 py-3 text-[17px] font-semibold transition-colors"
        :class="index === active ? 'border-neutral-800 text-neutral-900 dark:border-neutral-100 dark:text-neutral-100' : 'border-transparent text-neutral-500 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100'"
        @click="active = index"
      >
        {{ shot.title }}
      </button>
    </div>

    <div :id="`shot-panel-${current.name}`" role="tabpanel" :aria-labelledby="`shot-tab-${current.name}`" class="pt-7">
      <p class="max-w-4xl text-lg/relaxed text-neutral-700 dark:text-neutral-300">
        {{ current.note }}
      </p>

      <div class="mt-7 grid items-start gap-7 lg:grid-cols-2">
        <figure v-for="side in sides" :key="side.key" class="m-0 min-w-0">
          <figcaption class="mb-3 flex items-center justify-between gap-3 text-[15px]">
            <span class="flex items-center gap-2.5 text-lg font-semibold text-neutral-900 dark:text-neutral-100">
              <ComparePlatformIcon :name="side.name" :icon="side.icon" :lucity="side.lucity" />
              {{ side.name }}
            </span>
            <span class="text-neutral-600 dark:text-neutral-400">{{ caption(side.capture) }}</span>
          </figcaption>

          <video
            v-if="side.capture?.kind === 'video'"
            :key="side.capture.file"
            :src="side.capture.file"
            :poster="side.capture.poster"
            width="1280"
            height="800"
            controls
            playsinline
            preload="none"
            class="block aspect-[16/10] h-auto w-full rounded-xl border border-default bg-stone-200 shadow-[0_12px_32px_-18px_rgb(48_28_14/0.45)] dark:bg-stone-800"
          />
          <img
            v-else-if="side.capture"
            :key="side.capture.file"
            :src="side.capture.file"
            :alt="`${side.name}: ${current.title}`"
            width="1280"
            height="800"
            loading="lazy"
            class="block aspect-[16/10] h-auto w-full rounded-xl border border-default bg-stone-200 object-cover object-top shadow-[0_12px_32px_-18px_rgb(48_28_14/0.45)] dark:bg-stone-800"
          >
          <div
            v-else
            class="flex aspect-[16/10] w-full flex-col items-center justify-center gap-1 rounded-xl border border-dashed border-stone-400 p-6 text-center dark:border-stone-600"
          >
            <span class="font-semibold text-neutral-800 dark:text-neutral-200">{{ side.name }}: {{ current.title }}</span>
            <span class="text-[15px] text-neutral-600 dark:text-neutral-400">This screen hasn't been captured yet.</span>
          </div>
        </figure>
      </div>
    </div>
  </div>
</template>
