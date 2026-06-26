<script setup lang="ts">
const appUrl = 'https://lucity.cloud/app';
const signInUrl = `${appUrl}/login`;
const signUpUrl = `${appUrl}/login`;
</script>

<template>
  <header class="sticky top-0 z-50 bg-stone-100/85 px-6 pt-6 pb-4 backdrop-blur-md">
    <div class="grid grid-cols-2 items-center md:grid-cols-3">
      <NuxtLink to="/" class="flex items-center gap-2">
        <img src="/logo-light.svg" alt="Lucity" width="36" height="36" class="h-9 w-9 shrink-0">
        <span class="wordmark text-3xl leading-none text-neutral-900">Lucity</span>
      </NuxtLink>

      <p class="hidden justify-center text-lg text-neutral-600 md:flex">
        Built in Switzerland, hosted in the EU
      </p>

      <div class="flex items-center justify-end gap-6">
        <a :href="signInUrl" class="text-lg text-neutral-600 transition-colors hover:text-neutral-900">
          Sign in
        </a>
        <a
          :href="signUpUrl"
          class="rounded-lg bg-teal-500 px-5 py-2.5 text-lg font-semibold text-white transition-colors hover:bg-teal-600"
        >
          Sign up
        </a>
      </div>
    </div>

    <p class="mt-2 hidden text-right text-lg text-neutral-600 md:block">
      14-days free trial with €5 credits
    </p>
  </header>
</template>

<style scoped>
.wordmark {
  font-family: 'Redaction 35', Georgia, serif;
}
</style>
