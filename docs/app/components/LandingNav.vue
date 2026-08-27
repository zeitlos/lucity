<script setup lang="ts">
const appUrl = 'https://lucity.cloud/app';
const signInUrl = `${appUrl}/login`;
const signUpUrl = `${appUrl}/login`;
</script>

<template>
  <header class="sticky top-0 z-50 px-6 pt-6 pb-4">
    <div class="nav-blur" aria-hidden="true">
      <div class="nav-blur__layer" />
      <div class="nav-blur__layer" />
      <div class="nav-blur__layer" />
      <div class="nav-blur__tint" />
    </div>

    <div class="relative grid grid-cols-3 items-center xl:grid-cols-3">
      <NuxtLink to="/" class="flex items-center gap-2">
        <img src="/logo-light.svg" alt="Lucity" width="36" height="36" class="h-9 w-9 shrink-0 dark:hidden">
        <img src="/logo-dark.svg" alt="Lucity" width="36" height="36" class="hidden h-9 w-9 shrink-0 dark:block">
        <span class="wordmark text-3xl leading-none text-neutral-900 dark:text-neutral-100">Lucity</span>
      </NuxtLink>

      <p class="hidden justify-center text-lg text-neutral-600 xl:flex dark:text-neutral-400">
        Built in Switzerland, hosted in the EU
      </p>

      <div class="flex items-center justify-end gap-4 md:gap-6 col-span-2 xl:col-span-1">
        <a :href="signInUrl" class="hidden text-lg text-neutral-600 transition-colors hover:text-neutral-900 sm:inline dark:text-neutral-400 dark:hover:text-neutral-100">
          Sign in
        </a>
        <a
          :href="signUpUrl"
          class="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-white shadow-sm transition-colors hover:bg-primary-600 sm:gap-3 sm:px-5"
        >
          <span class="text-center leading-tight">
            <span class="block text-lg font-semibold">Sign up</span>
            <span class="hidden text-sm text-white/80 sm:block">14-days free trial with CHF 5 credits</span>
          </span>
          <UIcon name="i-lucide-arrow-right" class="size-5 shrink-0" />
        </a>
      </div>
    </div>
  </header>
</template>

<style scoped>
.wordmark {
  font-family: 'Redaction 35', Georgia, serif;
}

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
