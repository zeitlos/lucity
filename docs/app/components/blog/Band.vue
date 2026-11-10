<template>
  <section class="band">
    <slot />
  </section>
</template>

<style scoped>
.band {
  background: oklch(0.975 0.012 80);
  box-shadow: inset 0 1px 0 oklch(0.91 0.015 80), inset 0 -1px 0 oklch(0.91 0.015 80);
}

.dark .band {
  background: oklch(0.17 0.025 55);
  box-shadow: inset 0 1px 0 oklch(1 0 0 / .06), inset 0 -1px 0 oklch(1 0 0 / .06);
}
</style>
