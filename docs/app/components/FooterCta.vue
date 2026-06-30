<script setup lang="ts">
const appUrl = 'https://lucity.cloud/app';
</script>

<template>
  <section class="px-6">
    <div class="mx-auto flex min-h-[34rem] max-w-wide flex-col overflow-hidden rounded-[15px] px-6 py-12 sm:px-10 sm:py-14 md:min-h-[44rem] md:px-20 md:py-20 bg-[linear-gradient(to_bottom,#2f54a3_0%,#29b9fc_49%,#f6fcff_100%)]">
      <div class="max-w-xl space-y-8 font-display text-4xl leading-tight text-white sm:text-5xl md:text-6xl">
        <p>Don’t hire a half-ass DevOps Team.</p>
        <p>Get a whole-ass platform that does it for you.</p>
      </div>

      <a
        :href="`${appUrl}/login`"
        class="mt-auto self-end rounded-[15px] bg-[#00cf85] px-7 py-4 text-xl font-semibold text-white transition-colors hover:bg-[#00b876]"
      >
        Deploy your first App
      </a>
    </div>
  </section>
</template>
