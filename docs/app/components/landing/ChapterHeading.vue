<script setup lang="ts">
const props = defineProps<{
  id: string;
  beat: 'joy' | 'grow' | 'keep';
  promise: string;
  rest: string;
}>();

const beats = ['joy', 'grow', 'keep'] as const;

const beatColor: Record<typeof beats[number], string> = {
  joy: 'oklch(0.55 0.15 300)',
  grow: 'oklch(0.58 0.17 35)',
  keep: 'oklch(0.58 0.20 0)',
};

function stroke(name: typeof beats[number]) {
  return name === props.beat ? beatColor[name] : 'var(--beat-off)';
}
</script>

<template>
  <div class="flex flex-wrap items-end justify-between gap-x-18 gap-y-6">
    <div class="min-w-0 flex-[1_1_560px]">
      <div class="beats mb-5.5 flex gap-1.5" aria-hidden="true">
        <svg v-for="name in beats" :key="name" width="34" height="10" viewBox="0 0 34 10" fill="none" :stroke="stroke(name)" stroke-width="2.5" stroke-linecap="round"><path d="M2 6c5-6 10 6 15 0s10 6 15 0" /></svg>
      </div>
      <h2 :id="id" class="font-display text-4xl/11 text-balance text-stone-900 md:text-5xl/14 lg:text-6xl/17 dark:text-stone-100">
        <span class="relative whitespace-nowrap">{{ promise }}<img :src="`/img/underline-${beat}.svg`" alt="" class="pointer-events-none absolute bottom-0 left-0 h-[0.32em] w-full"></span>,<br>
        <span class="text-stone-500">{{ rest }}</span>
      </h2>
    </div>
    <p class="max-w-[540px] flex-[1_1_380px] text-xl leading-relaxed text-pretty text-stone-700 dark:text-stone-300">
      <slot />
    </p>
  </div>
</template>

<style scoped>
.beats {
  --beat-off: oklch(0.82 0.03 75);
}

.dark .beats {
  --beat-off: oklch(0.40 0.03 55);
}
</style>
