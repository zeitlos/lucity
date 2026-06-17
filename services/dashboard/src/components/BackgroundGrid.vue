<template>
  <div class="bg-grid" aria-hidden="true" />
</template>

<style scoped>
.bg-grid {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background-image:
    radial-gradient(circle 1.5px at 0 0, oklch(0.55 0.03 80 / 0.25) 1.5px, transparent 1.5px),
    linear-gradient(oklch(0.60 0.02 80 / 0.12) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.60 0.02 80 / 0.12) 1px, transparent 1px);
  background-size: 60px 60px;
  mask-image: linear-gradient(180deg, black 0%, rgb(0 0 0 / 0.5) 30%, transparent 55%);
  -webkit-mask-image: linear-gradient(180deg, black 0%, rgb(0 0 0 / 0.5) 30%, transparent 55%);
}

.bg-grid::after {
  content: '';
  position: fixed;
  inset: 0;
  background-image:
    linear-gradient(oklch(0.60 0.02 80 / 0.06) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.60 0.02 80 / 0.06) 1px, transparent 1px);
  background-size: 12px 12px;
  mask-image: linear-gradient(180deg, black 0%, rgb(0 0 0 / 0.3) 20%, transparent 40%);
  -webkit-mask-image: linear-gradient(180deg, black 0%, rgb(0 0 0 / 0.3) 20%, transparent 40%);
}

:global(.dark) .bg-grid {
  background-image:
    radial-gradient(circle 1.5px at 0 0, oklch(0.80 0.03 80 / 0.20) 1.5px, transparent 1.5px),
    linear-gradient(oklch(0.80 0.02 80 / 0.08) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.80 0.02 80 / 0.08) 1px, transparent 1px);
}
:global(.dark) .bg-grid::after {
  background-image:
    linear-gradient(oklch(0.80 0.02 80 / 0.04) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.80 0.02 80 / 0.04) 1px, transparent 1px);
}
</style>
