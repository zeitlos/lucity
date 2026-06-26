<script setup lang="ts">
useHead({
  link: [
    { rel: 'stylesheet', type: 'text/css', href: 'https://cdn.jsdelivr.net/gh/devicons/devicon@latest/devicon.min.css' },
  ],
});
</script>

<template>
  <div class="min-h-screen bg-stone-100 font-sans text-lg text-neutral-800 dark:bg-stone-950 dark:text-neutral-200">
    <LandingNav />
    <slot />
  </div>
</template>
