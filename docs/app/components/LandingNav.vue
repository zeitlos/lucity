<script setup lang="ts">
const appUrl = 'https://lucity.cloud/app';
const signInUrl = `${appUrl}/login`;
const signUpUrl = `${appUrl}/login`;
</script>

<template>
  <header class="sticky top-0 z-50 bg-stone-100/85 px-6 pt-6 pb-4 backdrop-blur-md dark:bg-stone-950/85">
    <div class="grid grid-cols-2 items-center md:grid-cols-3">
      <NuxtLink to="/" class="flex items-center gap-2">
        <img src="/logo-light.svg" alt="Lucity" width="36" height="36" class="h-9 w-9 shrink-0 dark:hidden">
        <img src="/logo-dark.svg" alt="Lucity" width="36" height="36" class="hidden h-9 w-9 shrink-0 dark:block">
        <span class="wordmark text-3xl leading-none text-neutral-900 dark:text-neutral-100">Lucity</span>
      </NuxtLink>

      <p class="hidden justify-center text-lg text-neutral-600 md:flex dark:text-neutral-400">
        Built in Switzerland, hosted in the EU
      </p>

      <div class="flex items-center justify-end gap-4 md:gap-6">
        <UColorModeButton class="text-neutral-600 dark:text-neutral-400" />
        <a :href="signInUrl" class="text-lg text-neutral-600 transition-colors hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100">
          Sign in
        </a>
        <a
          :href="signUpUrl"
          class="flex items-center gap-3 rounded-lg bg-teal-500 px-5 py-2 text-white shadow-sm transition-colors hover:bg-teal-600 dark:bg-teal-400 dark:hover:bg-teal-500"
        >
          <span class="text-center leading-tight">
            <span class="block text-lg font-semibold">Sign up</span>
            <span class="block text-sm text-white/80">14-days free trial with €5 credits</span>
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
</style>
