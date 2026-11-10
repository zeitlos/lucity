<script setup lang="ts">
const ways = ['Push to deploy', 'Ask your agent', 'Script it'];
const way = ref(0);

const features = [
  { name: 'Push to deploy', text: 'Connect a GitHub repo. Lucity detects the framework, builds it and ships every push.' },
  { name: 'A live URL, then your domain', text: "Every service gets an address with HTTPS on its first deploy. Add your own domain when you're ready." },
  { name: 'Changes in seconds', text: 'Variables, resources and the start command apply without a rebuild.' },
  { name: 'Your data, right there', text: 'Browse tables and run SQL next to your services, with the slowest queries one tab away.' },
];
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="joy">
    <LandingChapterHeading id="joy" beat="joy" promise="A joy to use" rest="from the first push.">
      From repository to running app in a few clicks, and every change after that is just as quick. Push, ask your agent or script it. It is the same platform underneath.
    </LandingChapterHeading>

    <LandingStage tint="joy" class="joy-stage mt-14 px-4 pt-7 pb-10 sm:px-9">
      <LandingTabs v-model="way" :options="ways" label="Ways to deploy" id-prefix="ways" />
      <div id="ways-panel" role="tabpanel" :aria-labelledby="`ways-tab-${way}`">
        <LandingPushScene v-if="way === 0" />
        <LandingAgentScene v-else-if="way === 1" />
        <LandingScriptScene v-else />
      </div>
    </LandingStage>

    <div class="mt-10 grid grid-cols-[repeat(auto-fit,minmax(min(250px,100%),1fr))] gap-x-10 border-b border-stone-300 dark:border-stone-800">
      <div v-for="feature in features" :key="feature.name" class="border-t border-stone-300 py-5.5 dark:border-stone-800">
        <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ feature.name }}</b>
        <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ feature.text }}</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.joy-stage :deep(.tab) {
  --tab-off: oklch(0.38 0.05 300);
}
</style>
