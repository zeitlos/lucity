<script setup lang="ts">
definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

useSeo({
  title: 'Lucity: The PaaS you can leave',
  description: 'Open-source PaaS on Kubernetes. Git push to deploy, environments out of the box, and a real exit door. Eject to standard Helm charts at any time. Self-hostable, no lock-in.',
  type: 'website',
});
</script>

<template>
  <section class="mx-auto px-6 mt-48 max-[650px]:mt-12 flex flex-col gap-16 min-[1800px]:grid min-[1800px]:grid-cols-12 min-[1800px]:items-end min-[1800px]:gap-24">
    <h1 class="font-display text-7xl/22 max-[650px]:text-4xl/[1.15] text-neutral-800 dark:text-neutral-100 min-[1800px]:col-span-5 min-[1800px]:mb-24">
      The Deployment Platform<br>
      that is a <span class="squiggle squiggle-joy">joy to use</span>,<br>
      <span class="squiggle squiggle-grow">built to grow</span>, and always<br>
      <span class="squiggle squiggle-keep">yours to keep</span>.
    </h1>

    <HeroDemo class="w-full min-[1800px]:col-span-7" />
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
