<script setup lang="ts">
import tech from '~/assets/landing/tech.json';
import releases from '~/assets/landing/releases.json';

const emit = defineEmits<{ search: [] }>();

const read = [
  { to: '/docs/quickstart', icon: 'i-lucide-book-open', title: 'Docs', text: 'A quickstart, then every part of the platform in plain words.' },
  { to: '/blog', icon: 'i-lucide-pen-line', title: 'Blog', text: 'Essays and news from building Lucity.' },
  { to: 'https://github.com/zeitlos/lucity/releases', icon: 'i-lucide-scroll-text', title: 'Changelog', text: 'What shipped, release by release.' },
];

function techPath(label: string) {
  return tech.find(item => item.label === label)?.d;
}

const guides = [
  { to: '/docs/guides/nextjs', label: 'Next.js', path: techPath('Next.js') },
  { to: '/docs/guides/nuxt', label: 'Nuxt', path: techPath('Nuxt') },
  { to: '/docs/guides/sveltekit', label: 'SvelteKit', path: techPath('SvelteKit') },
  { to: '/docs/guides/astro', label: 'Astro', path: techPath('Astro') },
  { to: '/docs/guides/express', label: 'Express', icon: 'i-simple-icons-express' },
  { to: '/docs/guides/fastapi', label: 'FastAPI', path: techPath('FastAPI') },
  { to: '/docs/guides/django', label: 'Django', path: techPath('Django') },
  { to: '/docs/guides/laravel', label: 'Laravel', path: techPath('Laravel') },
  { to: '/docs/guides/go', label: 'Go', path: techPath('Go') },
  { to: '/docs/self-hosting', label: 'Self-hosting', icon: 'i-lucide-server' },
];

const shipped = releases
  .flatMap(release => release.items.map(item => ({ date: release.date, name: item.name })))
  .slice(0, 3);

const { data: post } = await useAsyncData('nav-latest-post', () =>
  queryCollection('blog').order('date', 'DESC').select('path', 'title', 'date').first(),
);

const postDate = computed(() =>
  post.value?.date
    ? new Date(post.value.date).toLocaleDateString('en-GB', { day: 'numeric', month: 'long', timeZone: 'UTC' })
    : '',
);
</script>

<template>
  <NavPanel>
    <div>
      <p class="mb-1.5 px-3 text-sm text-stone-600 dark:text-stone-400">Read</p>
      <NavRow v-for="item in read" :key="item.to" v-bind="item" tint="orange" />
    </div>
    <div>
      <p class="mb-1.5 px-3 text-sm text-stone-600 dark:text-stone-400">Guides</p>
      <div class="grid grid-cols-2 gap-x-1">
        <NavMini v-for="item in guides" :key="item.to" v-bind="item" />
      </div>
    </div>

    <template #foot>
      <button
        type="button"
        class="-my-1.5 -ml-2.5 inline-flex min-h-10 cursor-pointer items-center gap-2.5 rounded-[10px] pr-3 pl-2.5 text-stone-700 transition-colors hover:bg-stone-900/5 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent) dark:text-stone-400 dark:hover:bg-white/6"
        @click="emit('search')"
      >
        <UIcon name="i-lucide-search" class="size-4.5" />
        <span>Search the docs</span>
        <kbd class="rounded-md px-[7px] py-0.5 font-sans text-[13px] text-stone-600 inset-ring inset-ring-stone-300 dark:text-stone-400 dark:inset-ring-stone-700">⌘K</kbd>
      </button>
      <NuxtLink
        to="https://github.com/zeitlos/lucity"
        class="rounded-sm text-[15px] font-semibold text-stone-900 underline decoration-1 underline-offset-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent) dark:text-stone-100"
      >
        Lucity on GitHub
      </NuxtLink>
    </template>

    <template #stage>
      <NavStage :to="post?.path ?? '/blog'" tint="grow" :action="post ? 'Read the post' : 'Open the blog'">
        <template v-if="post">
          <span class="text-sm text-[oklch(0.40_0.06_60)] dark:text-[oklch(0.80_0.06_66)]">Latest post, {{ postDate }}</span>
          <span class="mt-2 font-display text-[22px]/[1.18]">{{ post.title }}</span>
        </template>
        <span class="mt-4.5 mb-1 border-t border-[oklch(0.25_0.04_55/.14)] dark:border-white/15 pt-3.5 text-sm text-[oklch(0.40_0.06_60)] dark:text-[oklch(0.80_0.06_66)]">Recently shipped</span>
        <span
          v-for="item in shipped"
          :key="item.name"
          class="grid grid-cols-[46px_minmax(0,1fr)] gap-x-2.5 py-[7px] text-[14.5px]/[1.35]"
        >
          <span class="font-semibold">{{ item.date }}</span>
          <span>{{ item.name }}</span>
        </span>
      </NavStage>
    </template>
  </NavPanel>
</template>
