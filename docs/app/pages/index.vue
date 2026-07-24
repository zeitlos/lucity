<script setup lang="ts">
definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

useSeo({
  title: 'Deploy anything. Own your stack.',
  description: 'Get your apps online and running in production, without the lock-in. Lucity is the European open source alternative to Vercel, Heroku, and Railway: the same effortless developer experience, but your whole stack stays yours to move anywhere, anytime.',
  type: 'website',
});
</script>

<template>
  <section class="mx-auto px-6 mt-48 max-sm:mt-12 flex flex-col gap-16 3xl:grid 3xl:grid-cols-12 3xl:items-end 3xl:gap-24">
    <h1 class="font-display text-4xl/[1.15] sm:text-5xl/[1.1] md:text-6xl/[1.1] lg:text-7xl/22 text-neutral-800 dark:text-neutral-100 3xl:col-span-5 3xl:mb-24">
      The Deployment Platform<br>
      that is a <span class="squiggle squiggle-joy">joy to use</span>,<br>
      <span class="squiggle squiggle-grow">built to grow</span>, and always<br>
      <span class="squiggle squiggle-keep">yours to keep</span>.
    </h1>

    <HeroDemo class="w-full 3xl:col-span-7" />
  </section>

  <TechCarousel class="mt-32" />

  <FounderLetter class="mt-32" />

  <TestimonialQuote class="mt-40" />

  <DeploymentTrilemma class="mt-40" />

  <LucityBend class="mt-12" />

  <BentoGrid class="mt-40" />

  <FooterCta class="mt-40" />

  <SiteFooter class="mt-24" />
</template>

<style scoped>
.squiggle {
  position: relative;
  white-space: nowrap;
}

.squiggle::after {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 0.34em;
  background-repeat: no-repeat;
  background-position: center;
  background-size: 100% 100%;
  clip-path: inset(0 100% 0 0);
  animation: highlighter 0.55s ease-out forwards;
}

.squiggle-joy::after {
  background-image: url('/img/underline-joy.svg');
  animation-delay: 0.35s;
}

.squiggle-grow::after {
  background-image: url('/img/underline-grow.svg');
  animation-delay: 0.9s;
}

.squiggle-keep::after {
  background-image: url('/img/underline-keep.svg');
  animation-delay: 1.45s;
}

@keyframes highlighter {
  to {
    clip-path: inset(0 0 0 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .squiggle::after {
    animation: none;
    clip-path: inset(0 0 0 0);
  }
}
</style>
