<script setup lang="ts">
import tech from '~/assets/landing/tech.json';

const label = `${tech.filter(item => item.label !== 'Docker').map(item => item.label).join(', ')} and any public container image`;
</script>

<template>
  <section aria-label="Supported technology">
    <p class="px-6 text-center text-lg text-stone-700 dark:text-stone-400">
      Runs what you already write. No Dockerfile needed.
    </p>
    <div class="marquee mt-9" role="img" :aria-label="label">
      <div class="marquee-track">
        <div v-for="copy in 2" :key="copy" class="marquee-group" :aria-hidden="copy === 2 ? 'true' : undefined">
          <span
            v-for="item in tech"
            :key="item.label"
            class="flex items-center gap-3 text-xl font-semibold whitespace-nowrap text-stone-800 dark:text-stone-200"
          >
            <svg class="size-7.5 shrink-0" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path :d="item.d" /></svg>
            {{ item.label }}
          </span>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.marquee {
  overflow: hidden;
  mask-image: linear-gradient(to right, transparent, #000 7%, #000 93%, transparent);
}

.marquee-track {
  display: flex;
  width: max-content;
  animation: marquee 70s linear infinite;
}

.marquee:hover .marquee-track {
  animation-play-state: paused;
}

.marquee-group {
  display: flex;
  gap: 60px;
  padding-right: 60px;
}

@keyframes marquee {
  to {
    transform: translateX(-50%);
  }
}

@media (prefers-reduced-motion: reduce) {
  .marquee-track {
    animation: none;
  }
}
</style>
