<script setup lang="ts">
const appUrl = '/app';
const signInUrl = `${appUrl}/login`;
const signUpUrl = `${appUrl}/login`;
const githubUrl = 'https://github.com/zeitlos/lucity';

const links = [
  { label: 'Docs', to: '/docs/quickstart' },
  { label: 'Blog', to: '/blog' },
  { label: 'Pricing', to: '/pricing' },
];

const tucked = ref(false);
const menuOpen = ref(false);

function onScroll() {
  tucked.value = tucked.value ? window.scrollY > 24 : window.scrollY > 96;
}

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true });
  onScroll();
});

onBeforeUnmount(() => window.removeEventListener('scroll', onScroll));
</script>

<template>
  <header
    class="sticky top-0 z-50 px-6 pt-6 pb-4"
    :class="{ 'max-lg:bg-stone-100 max-lg:dark:bg-stone-950': menuOpen }"
  >
    <div class="nav-blur" aria-hidden="true">
      <div class="nav-blur__layer" />
      <div class="nav-blur__layer" />
      <div class="nav-blur__layer" />
      <div class="nav-blur__tint" />
    </div>

    <div class="relative grid grid-cols-[auto_1fr] items-center gap-4 lg:grid-cols-3 lg:gap-0">
      <NuxtLink to="/" class="flex items-center">
        <LandingLogo :tucked="tucked" class="text-2xl text-neutral-900 dark:text-neutral-100" />
      </NuxtLink>

      <nav aria-label="Main" class="hidden justify-center gap-9 lg:flex">
        <NuxtLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="text-lg text-neutral-600 transition-colors hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100"
        >
          {{ link.label }}
        </NuxtLink>
      </nav>

      <div class="flex items-center justify-end gap-2 lg:gap-5">
        <a
          :href="githubUrl"
          target="_blank"
          rel="noopener"
          aria-label="Lucity on GitHub"
          class="-mr-1 hidden size-10 items-center justify-center text-neutral-600 transition-colors hover:text-neutral-900 lg:flex dark:text-neutral-400 dark:hover:text-neutral-100"
        >
          <UIcon name="i-simple-icons-github" class="size-5.5" />
        </a>
        <a :href="signInUrl" class="hidden text-lg text-neutral-600 transition-colors hover:text-neutral-900 lg:inline dark:text-neutral-400 dark:hover:text-neutral-100">
          Sign in
        </a>
        <a
          :href="signUpUrl"
          class="rounded-lg bg-primary px-3.5 py-2 text-base leading-6 font-semibold text-white shadow-sm transition-colors hover:bg-primary-600 sm:px-4.5 sm:text-lg sm:leading-6"
        >
          Get started
        </a>
        <button
          type="button"
          :aria-label="menuOpen ? 'Close menu' : 'Open menu'"
          :aria-expanded="menuOpen"
          aria-controls="landing-menu"
          class="-mr-2.5 flex size-11 cursor-pointer items-center justify-center rounded-lg text-neutral-800 lg:hidden dark:text-neutral-100"
          @click="menuOpen = !menuOpen"
        >
          <UIcon :name="menuOpen ? 'i-lucide-x' : 'i-lucide-menu'" class="size-6" />
        </button>
      </div>
    </div>

    <div
      v-if="menuOpen"
      id="landing-menu"
      class="absolute inset-x-0 top-full h-dvh border-t border-stone-200 bg-stone-100 px-6 pt-2 lg:hidden dark:border-stone-800 dark:bg-stone-950"
    >
      <nav aria-label="Main" class="flex flex-col">
        <NuxtLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="flex h-16 items-center justify-between border-b border-stone-200 text-xl text-neutral-800 dark:border-stone-800 dark:text-neutral-100"
        >
          {{ link.label }}
          <UIcon name="i-lucide-chevron-right" class="size-5 text-stone-400" />
        </NuxtLink>
        <a
          :href="githubUrl"
          target="_blank"
          rel="noopener"
          class="flex h-16 items-center justify-between border-b border-stone-200 text-xl text-neutral-800 dark:border-stone-800 dark:text-neutral-100"
        >
          GitHub
          <UIcon name="i-lucide-arrow-up-right" class="size-5 text-stone-400" />
        </a>
      </nav>

      <div class="mt-8 flex flex-col gap-3">
        <a
          :href="signInUrl"
          class="rounded-lg border border-stone-300 bg-white/50 px-5 py-3.5 text-center text-lg font-semibold text-neutral-800 dark:border-stone-700 dark:bg-white/5 dark:text-neutral-100"
        >
          Sign in
        </a>
        <a
          :href="signUpUrl"
          class="rounded-lg border border-transparent bg-primary px-5 py-3.5 text-center text-lg font-semibold text-white transition-colors hover:bg-primary-600"
        >
          Get started
        </a>
        <p class="mt-1 text-center text-base text-stone-600 dark:text-stone-400">
          14-day free trial with CHF 5 in credits
        </p>
      </div>
    </div>
  </header>
</template>

<style scoped>
/* iOS-style progressive blur: stacked backdrop layers, each masked to a band,
   so the blur is strongest at the top and fades to nothing below the bar —
   no hard edge. Extends past the header bottom so the fade trails off. */
.nav-blur {
  position: absolute;
  inset: 0 0 -2rem 0;
  z-index: 0;
  pointer-events: none;
}

.nav-blur__layer {
  position: absolute;
  inset: 0;
}

.nav-blur__layer:nth-child(1) {
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  -webkit-mask-image: linear-gradient(to bottom, #000 0%, #000 33%, transparent 66%);
  mask-image: linear-gradient(to bottom, #000 0%, #000 33%, transparent 66%);
}

.nav-blur__layer:nth-child(2) {
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  -webkit-mask-image: linear-gradient(to bottom, transparent 16%, #000 50%, transparent 83%);
  mask-image: linear-gradient(to bottom, transparent 16%, #000 50%, transparent 83%);
}

.nav-blur__layer:nth-child(3) {
  backdrop-filter: blur(2px);
  -webkit-backdrop-filter: blur(2px);
  -webkit-mask-image: linear-gradient(to bottom, transparent 50%, #000 83%, transparent 100%);
  mask-image: linear-gradient(to bottom, transparent 50%, #000 83%, transparent 100%);
}

.nav-blur__tint {
  position: absolute;
  inset: 0;
  background: linear-gradient(to bottom, var(--ui-bg) 0%, transparent 100%);
  opacity: 0.7;
}
</style>
