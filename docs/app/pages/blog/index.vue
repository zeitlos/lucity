<script setup lang="ts">
import { authorOf, formatDate, summarize } from '~/components/blog/post';
import { releasesUrl } from '~/components/blog/releases';

definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

useSeo({
  title: 'Blog',
  description: 'Updates, deep dives, and stories from the Lucity team.',
  type: 'website',
});

interface NavigationNode {
  title: string;
  path: string;
  children?: NavigationNode[];
}

function guidesIn(nodes: NavigationNode[]): NavigationNode[] {
  for (const node of nodes) {
    if (node.path === '/docs/guides' && node.children) {
      return node.children.filter(child => child.path !== node.path);
    }
    const nested = guidesIn(node.children ?? []);
    if (nested.length) return nested;
  }
  return [];
}

const { data } = await useAsyncData('blog-index', async () => {
  const [posts, navigation] = await Promise.all([
    queryCollection('blog').order('date', 'DESC').all(),
    queryCollectionNavigation('docs'),
  ]);
  return {
    posts: posts.map(summarize),
    guides: guidesIn(navigation as NavigationNode[]).map(guide => ({ path: guide.path, title: guide.title })),
  };
});

const posts = computed(() => data.value?.posts ?? []);
const guides = computed(() => data.value?.guides ?? []);

const filters = ['Everything', 'Essays', 'Releases', 'Guides'];
const filter = ref(0);

const featured = computed(() => filter.value === 0 ? posts.value[0] : undefined);
const essays = computed(() => {
  if (filter.value === 1) return posts.value;
  return filter.value === 0 ? posts.value.slice(1) : [];
});
const showReleases = computed(() => filter.value === 0 || filter.value === 2);

const starters = [
  { to: '/comparisons/vs-railway', name: 'Lucity vs Railway', text: 'The same deploy experience, on a platform you can take with you.' },
  { to: '/comparisons/vs-heroku', name: 'Lucity vs Heroku', text: 'What changed since git push heroku main, and what did not.' },
  { to: '/comparisons/vs-vercel', name: 'Lucity vs Vercel', text: 'For applications that need more than a frontend.' },
  { to: '/docs/quickstart', name: 'The quickstart', text: 'From a fresh account to a live app with a database, in one sitting.' },
];
</script>

<template>
  <div>
    <section class="mx-auto max-w-content px-6 pt-10 sm:pt-26" aria-labelledby="top">
      <div class="flex flex-wrap items-end justify-between gap-x-18 gap-y-6">
        <div class="min-w-0 flex-[1_1_560px]">
          <h1 id="top" class="load-in font-display text-4xl/[1.15] text-stone-900 sm:text-5xl/[1.1] md:text-6xl/[1.1] lg:text-7xl/[1.1] dark:text-stone-100" style="--load-order: 0">
            Notes from <span class="relative whitespace-nowrap">building Lucity<img src="/img/underline-joy.svg" alt="" class="pointer-events-none absolute bottom-0 left-0 h-[0.32em] w-full"></span>.
          </h1>
        </div>
        <p class="load-in max-w-[540px] flex-[1_1_380px] text-xl leading-relaxed text-pretty text-stone-700 dark:text-stone-300" style="--load-order: 1">
          Essays on how software gets deployed and who ends up owning it, next to a running log of what shipped. Written by the people who build the platform.
        </p>
      </div>

      <div class="load-in mt-10 flex flex-wrap items-center justify-between gap-x-8 gap-y-3 sm:mt-14" style="--load-order: 2">
        <BlogFilterTabs v-model="filter" :options="filters" label="What to show" panel="blog-panel" />
        <a :href="releasesUrl" class="inline-flex min-h-11 items-center gap-2 text-base font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">
          <UIcon name="i-simple-icons-github" class="size-4.5 shrink-0" />
          Releases on GitHub
        </a>
      </div>
    </section>

    <div id="blog-panel" role="tabpanel" :aria-label="filters[filter]">
      <section v-if="featured" class="mx-auto mt-10 max-w-content px-6 sm:mt-14" aria-labelledby="featured">
        <div class="flex flex-wrap items-stretch gap-x-18 gap-y-10">
          <div class="flex min-w-0 flex-[1_1_440px] flex-col justify-center py-2">
            <p class="text-base text-stone-600 dark:text-stone-400">
              <time :datetime="featured.date">{{ formatDate(featured.date) }}</time> · {{ featured.minutes }} min read
            </p>
            <h2 id="featured" class="mt-4.5 font-display text-4xl/[1.12] text-balance text-stone-900 lg:text-5xl/[1.12] xl:text-[54px]/[1.12] dark:text-stone-100">
              <NuxtLink :to="featured.path" class="decoration-2 underline-offset-[6px] hover:underline">{{ featured.title }}</NuxtLink>
            </h2>
            <p v-if="featured.description" class="mt-5.5 max-w-[560px] text-xl leading-relaxed text-pretty text-stone-700 dark:text-stone-300">{{ featured.description }}</p>
            <BlogByline v-if="featured.author" :name="featured.author" class="mt-7.5">{{ authorOf(featured.author).role }}</BlogByline>
            <p class="mt-7.5 text-lg">
              <NuxtLink :to="featured.path" class="inline-flex min-h-11 items-center gap-2 font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">
                Read the essay
                <UIcon name="i-lucide-arrow-right" class="size-5 shrink-0" />
              </NuxtLink>
            </p>
          </div>

          <LandingStage tint="keep" class="art flex min-w-0 flex-[1.2_1_520px] flex-col justify-between gap-7 px-4 pt-8 sm:px-10 sm:pt-11" :class="{ 'pb-8 sm:pb-11': !featured.quote }">
            <img
              v-if="featured.image"
              :src="featured.image.src"
              :alt="featured.image.alt"
              class="m-auto block max-h-[420px] w-auto max-w-full rounded-xl"
            >
            <img
              v-else
              src="/img/landing/sketch-eject.svg"
              alt="A project on Lucity, and the same project as a copy running anywhere else, carried over as one zip"
              width="760"
              height="300"
              class="m-auto block h-auto w-full max-w-[700px] dark:hidden"
            >
            <img
              v-if="!featured.image"
              src="/img/landing/sketch-eject-night.svg"
              alt=""
              width="760"
              height="300"
              class="m-auto hidden h-auto w-full max-w-[700px] dark:block"
            >
            <p v-if="featured.quote" class="quote -mb-4.5 max-w-[560px] self-start bg-white px-5 pt-5 pb-9 font-display text-xl/[1.3] text-stone-900 sm:ml-3 sm:px-8 sm:pt-6.5 sm:pb-10 sm:text-[26px]/[1.3]">
              {{ featured.quote }}
            </p>
          </LandingStage>
        </div>
      </section>

      <section class="mx-auto max-w-content px-6" :class="featured ? 'mt-20 sm:mt-30' : 'mt-12 sm:mt-16'">
        <div class="flex flex-wrap items-start gap-x-22 gap-y-16">
          <div v-if="essays.length" class="min-w-0 flex-[1.5_1_560px]">
            <h2 class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">Essays</h2>
            <div class="mt-7 border-b border-stone-300 dark:border-stone-800">
              <BlogEssayRow v-for="post in essays" :key="post.path" :post="post" />
            </div>
          </div>

          <div v-if="showReleases" class="min-w-0 flex-[1_1_400px]">
            <h2 class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">Recently shipped</h2>
            <BlogReleaseLog :count="filter === 0 ? 3 : undefined" />
          </div>

          <div v-if="filter === 3" class="min-w-0 flex-[1_1_100%]">
            <h2 class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">Deploy guides</h2>
            <p class="mt-3.5 max-w-[640px] text-xl leading-relaxed text-pretty text-stone-700 dark:text-stone-300">Step by step, from an empty account to a running app on its own domain.</p>
            <BlogGuideLinks :guides="guides" class="mt-8" />
            <p class="mt-5 text-base">
              <NuxtLink to="/docs/quickstart" class="inline-flex min-h-11 items-center font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">Or start with the quickstart</NuxtLink>
            </p>
          </div>
        </div>
      </section>

      <template v-if="filter === 0">
        <BlogBand class="mt-20 py-16 sm:mt-34 sm:pt-24 sm:pb-22" aria-labelledby="start">
          <div class="mx-auto flex max-w-content flex-wrap items-start justify-between gap-x-18 gap-y-8 px-6">
            <div class="max-w-[460px] min-w-0 flex-[1_1_380px]">
              <h2 id="start" class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">New here? Start with these.</h2>
              <p class="mt-4 text-lg leading-relaxed text-pretty text-stone-700 dark:text-stone-300">The short answers to “how is this different from what I use today”, and the fastest way to see for yourself.</p>
            </div>
            <div class="min-w-0 flex-[1.6_1_520px] border-b border-stone-300 dark:border-stone-800">
              <NuxtLink v-for="starter in starters" :key="starter.to" :to="starter.to" class="group block border-t border-stone-300 py-5.5 dark:border-stone-800">
                <b class="block text-[19px] font-semibold text-stone-900 underline-offset-4 group-hover:underline dark:text-stone-100">{{ starter.name }}</b>
                <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ starter.text }}</span>
              </NuxtLink>
            </div>
          </div>
        </BlogBand>

        <section class="mx-auto mt-24 max-w-content px-6 sm:mt-38" aria-labelledby="guides">
          <div class="flex flex-wrap items-end justify-between gap-x-18 gap-y-4">
            <h2 id="guides" class="flex-[1_1_380px] font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">Deploy guides, one per framework.</h2>
            <p class="max-w-[540px] flex-[1_1_380px] text-lg leading-relaxed text-pretty text-stone-700 dark:text-stone-300">Step by step, from an empty account to a running app on its own domain.</p>
          </div>
          <BlogGuideLinks :guides="guides" class="mt-9" />
        </section>
      </template>
    </div>

    <BlogFollowAlong class="mt-24 sm:mt-38" />
  </div>
</template>

<style scoped>
.art {
  min-height: 480px;
}

.quote {
  box-shadow: 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
  transform: rotate(-.6deg);
}

@media (max-width: 640px) {
  .art {
    min-height: 0;
  }
}
</style>
