<script setup lang="ts">
const appUrl = '/app';
</script>

<template>
  <section class="px-6">
    <div class="mx-auto flex min-h-[34rem] max-w-wide flex-col overflow-hidden rounded-[15px] px-6 py-12 sm:px-10 sm:py-14 md:min-h-[44rem] md:px-20 md:py-20 bg-[linear-gradient(to_bottom,#2f54a3_0%,#29b9fc_49%,#f6fcff_100%)]">
      <div class="max-w-xl space-y-8 font-display text-4xl leading-tight text-white sm:text-5xl md:text-6xl">
        <p>Don’t hire a half-ass DevOps Team.</p>
        <p>Get a whole-ass platform that does it for you.</p>
      </div>

      <div class="mt-auto flex flex-col-reverse gap-3 self-end max-sm:self-stretch sm:flex-row sm:gap-4">
        <NuxtLink
          to="/docs/quickstart"
          class="flex items-center justify-center rounded-lg border border-white/70 bg-white/60 px-7 py-4 text-xl font-semibold text-neutral-900 backdrop-blur-sm transition-colors hover:bg-white"
        >
          Read the docs
        </NuxtLink>
        <a
          :href="`${appUrl}/login`"
          class="flex items-center justify-center gap-3 rounded-lg border border-transparent bg-primary-500 px-7 py-4 text-xl font-semibold text-white transition-colors hover:bg-primary-600"
        >
          Deploy your first app
          <UIcon name="i-lucide-arrow-right" class="size-5.5 shrink-0" />
        </a>
      </div>
    </div>
  </section>
</template>
