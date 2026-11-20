<script setup lang="ts">
defineProps<{
  to: string;
  tint: 'joy' | 'grow' | 'keep';
  action: string;
}>();
</script>

<template>
  <NuxtLink
    :to="to"
    class="nav-stage flex flex-col rounded-[14px] px-5.5 pt-5.5 pb-5 text-[oklch(0.25_0.03_55)] dark:text-stone-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent)"
    :class="`nav-stage-${tint}`"
  >
    <slot />
    <span class="mt-auto inline-flex items-center gap-2 pt-3.5 text-[15px] font-semibold">
      {{ action }}
      <UIcon name="i-lucide-arrow-right" class="nav-stage__arrow size-[17px]" />
    </span>
  </NuxtLink>
</template>

<style scoped>
.nav-stage {
  background-image: radial-gradient(circle at 1px 1px, oklch(0.25 0.04 55 / .14) 1px, transparent 0);
  background-size: 22px 22px;
  box-shadow: inset 0 0 0 1px oklch(0.25 0.04 55 / .07);
  transition: box-shadow 0.15s;
}

.nav-stage:hover {
  box-shadow: inset 0 0 0 1.5px oklch(0.25 0.04 55 / .28);
}

.nav-stage-joy {
  background-color: oklch(0.925 0.028 300);
}

.nav-stage-grow {
  background-color: oklch(0.93 0.038 68);
}

.nav-stage-keep {
  background-color: oklch(0.928 0.026 8);
}

.dark .nav-stage {
  background-image: radial-gradient(circle at 1px 1px, oklch(1 0 0 / .10) 1px, transparent 0);
  box-shadow: inset 0 0 0 1px oklch(1 0 0 / .09);
}

.dark .nav-stage:hover {
  box-shadow: inset 0 0 0 1.5px oklch(1 0 0 / .30);
}

.dark .nav-stage-joy {
  background-color: oklch(0.31 0.06 300);
}

.dark .nav-stage-grow {
  background-color: oklch(0.32 0.06 66);
}

.dark .nav-stage-keep {
  background-color: oklch(0.31 0.06 8);
}

.nav-stage__arrow {
  transition: translate 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.nav-stage:hover .nav-stage__arrow {
  translate: 3px 0;
}

@media (prefers-reduced-motion: reduce) {
  .nav-stage,
  .nav-stage__arrow {
    transition: none;
  }
}
</style>
