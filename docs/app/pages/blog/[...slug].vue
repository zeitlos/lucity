<script setup lang="ts">
import { kebabCase } from 'scule';
import BlogProseCode from '~/components/blog/ProseCode.vue';
import BlogProseLink from '~/components/blog/ProseLink.vue';
import { authorOf, formatDate, readingMinutes, splitPost, summarize } from '~/components/blog/post';

definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

const route = useRoute();

const { data: page } = await useAsyncData(kebabCase(route.path), () =>
  queryCollection('blog').path(route.path).first()
);

if (!page.value) {
  throw createError({ statusCode: 404, statusMessage: 'Post not found', fatal: true });
}

const { data: others } = await useAsyncData(`${kebabCase(route.path)}-more`, async () => {
  const posts = await queryCollection('blog').order('date', 'DESC').all();
  return posts.filter(post => post.path !== route.path).slice(0, 3).map(summarize);
});

const minutes = computed(() => readingMinutes(page.value?.body));
const author = computed(() => authorOf(page.value?.author));
const parts = computed(() => splitPost(page.value?.body));
const main = computed(() => ({ body: { type: 'minimark', value: parts.value.main } }));
const closing = computed(() => ({ body: { type: 'minimark', value: parts.value.closing } }));

const contents = computed(() => {
  const links = (page.value?.body as { toc?: { links?: { id: string; text: string }[] } } | undefined)?.toc?.links ?? [];
  return links.map(link => ({ id: link.id, text: link.text }));
});

const components = { a: BlogProseLink, pre: BlogProseCode };

useSeo({
  title: page.value.title,
  description: page.value.description,
  type: 'article',
  publishedAt: page.value.date,
});
</script>

<template>
  <div v-if="page">
    <article class="mx-auto max-w-content px-6 pt-6 sm:pt-20" aria-labelledby="top">
      <header class="max-w-[1040px]">
        <p class="text-base">
          <NuxtLink to="/blog" class="inline-flex min-h-11 items-center gap-2 text-stone-700 hover:text-stone-900 dark:text-stone-400 dark:hover:text-stone-100">
            <UIcon name="i-lucide-arrow-left" class="size-4.5 shrink-0" />
            Blog
          </NuxtLink>
        </p>
        <h1 id="top" class="load-in mt-3 font-display text-4xl/[1.15] text-stone-900 sm:mt-5 sm:text-5xl/[1.1] md:text-6xl/[1.1] lg:text-7xl/[1.1] dark:text-stone-100" style="--load-order: 0">
          {{ page.title }}
        </h1>
        <p v-if="page.description" class="load-in mt-5 max-w-[780px] text-xl/[1.5] text-pretty text-stone-800 sm:mt-7 sm:text-[23px]/[1.55] dark:text-stone-300" style="--load-order: 1">
          {{ page.description }}
        </p>
        <BlogByline :name="page.author" class="load-in mt-7 sm:mt-9" style="--load-order: 2">
          <time :datetime="page.date">{{ formatDate(page.date) }}</time> · {{ minutes }} min read
        </BlogByline>
      </header>

      <div class="post mt-8 border-t border-stone-300 pt-3 sm:mt-16 dark:border-stone-800">
        <BlogToc :links="contents" class="contents" />

        <div class="min-w-0">
          <ContentRenderer :value="main" :prose="false" :components="components" class="blog-prose" />

          <div v-if="page.author || parts.closing.length" class="sign mt-14 border-t border-stone-300 pt-8 sm:mt-18 sm:pt-9 dark:border-stone-800" :class="{ 'items-center': !parts.closing.length }">
            <BlogAvatar :name="page.author" large />
            <div class="min-w-0">
              <ContentRenderer v-if="parts.closing.length" :value="closing" :prose="false" :components="components" class="blog-prose blog-prose-closing" />
              <p v-if="page.author" class="text-lg/[1.6] text-stone-900 dark:text-stone-100">
                <b class="font-semibold">{{ page.author }}</b><template v-if="author.role">, <i class="text-stone-700 dark:text-stone-400">{{ author.role }}</i></template>
              </p>
            </div>
          </div>
        </div>
      </div>
    </article>

    <BlogBand class="mt-20 py-16 sm:mt-34 sm:pt-22 sm:pb-20" aria-labelledby="next">
      <div class="mx-auto max-w-content px-6">
        <h2 id="next" class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">Keep reading</h2>
        <div v-if="others?.length" class="mt-5 border-b border-stone-300 dark:border-stone-800">
          <BlogEssayRow v-for="post in others" :key="post.path" :post="post" arrow />
        </div>
        <p class="mt-5 text-base">
          <NuxtLink to="/blog" class="inline-flex min-h-11 items-center font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">All essays and releases</NuxtLink>
        </p>
      </div>
    </BlogBand>

    <BlogFollowAlong class="mt-24 sm:mt-38" />
  </div>
</template>

<style scoped>
.post {
  display: grid;
  grid-template-columns: minmax(0, 700px);
  gap: 20px 0;
}

.sign {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 0 16px;
}

@media (min-width: 640px) {
  .post {
    gap: 40px 0;
    padding-top: 24px;
  }

  .sign {
    gap: 0 20px;
  }
}

@media (min-width: 1100px) {
  .post {
    grid-template-columns: 232px minmax(0, 700px);
    gap: 0 112px;
    padding-top: 56px;
  }

  .contents {
    position: sticky;
    top: 112px;
    align-self: start;
    max-height: calc(100vh - 136px);
    overflow-y: auto;
  }
}
</style>

<style>
.blog-prose {
  --prose-ink: oklch(0.25 0.03 55);
  --prose-text: oklch(0.31 0.03 55);
  --prose-meta: oklch(0.47 0.04 55);
  --prose-rule: oklch(0.85 0.03 75);
  --prose-chip: oklch(0.25 0.04 55 / .07);
  --prose-dot: oklch(0.55 0.15 300);
  font-size: 18px;
  line-height: 1.7;
  color: var(--prose-text);
}

.dark .blog-prose {
  --prose-ink: oklch(0.95 0.02 80);
  --prose-text: oklch(0.86 0.025 75);
  --prose-meta: oklch(0.70 0.03 65);
  --prose-rule: oklch(0.30 0.03 55);
  --prose-chip: oklch(1 0 0 / .09);
  --prose-dot: oklch(0.72 0.12 300);
}

.blog-prose > :first-child {
  margin-top: 0;
}

.blog-prose p {
  margin: 24px 0 0;
  text-wrap: pretty;
}

.blog-prose strong,
.blog-prose b {
  font-weight: 650;
  color: var(--prose-ink);
}

.blog-prose a {
  color: var(--prose-ink);
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 4px;
  overflow-wrap: anywhere;
}

.blog-prose a:hover {
  text-decoration-thickness: 2px;
}

.blog-prose :is(h2, h3, h4) {
  scroll-margin-top: 7rem;
  color: var(--prose-ink);
  text-wrap: balance;
}

.blog-prose h2 {
  margin: 64px 0 0;
  font-family: var(--font-display);
  font-weight: 400;
  font-size: clamp(30px, 2.8vw, 40px);
  line-height: 1.15;
}

.blog-prose h2 + p {
  margin-top: 22px;
}

.blog-prose h3 {
  margin: 48px 0 0;
  font-family: var(--font-display);
  font-weight: 400;
  font-size: 26px;
  line-height: 1.2;
}

.blog-prose h4 {
  margin: 36px 0 0;
  font-size: 1em;
  font-weight: 650;
}

.blog-prose :is(h2, h3, h4) a {
  text-decoration: none;
}

.blog-prose :is(ul, ol) {
  margin: 22px 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.blog-prose li {
  position: relative;
  padding-left: 21px;
}

.blog-prose ul > li::before {
  content: "";
  position: absolute;
  left: 0;
  top: calc(0.85em - 3.5px);
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--prose-dot);
}

.blog-prose ol {
  counter-reset: step;
}

.blog-prose ol > li {
  counter-increment: step;
  padding-left: 34px;
}

.blog-prose ol > li::before {
  content: counter(step) ".";
  position: absolute;
  left: 0;
  font-weight: 650;
  color: var(--prose-ink);
}

.blog-prose li > :is(ul, ol) {
  margin-top: 10px;
}

.blog-prose li > p:first-child {
  margin-top: 0;
}

.blog-prose :not(pre) > code {
  font-family: var(--font-mono);
  font-size: .86em;
  padding: 2px 6px;
  border-radius: 5px;
  background: var(--prose-chip);
  color: var(--prose-ink);
}

.blog-prose blockquote {
  margin: 40px 0 0;
  padding: 30px 0 28px;
  border-top: 1px solid var(--prose-rule);
  border-bottom: 1px solid var(--prose-rule);
}

.blog-prose blockquote p {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(24px, 2.6vw, 34px);
  line-height: 1.28;
  color: var(--prose-ink);
}

.blog-prose blockquote p + p {
  margin-top: 16px;
}

.blog-prose blockquote + p {
  margin-top: 36px;
}

.blog-prose figure {
  margin: 40px 0 0;
}

.blog-prose figure img {
  display: block;
  box-sizing: border-box;
  width: 100%;
  height: auto;
  max-height: 580px;
  padding: 12px;
  object-fit: contain;
  border-radius: 16px;
  background: #f6ede0;
  box-shadow: inset 0 0 0 1px oklch(0.25 0.04 55 / .08);
}

.blog-prose figcaption {
  max-width: 700px;
  margin-top: 12px;
  font-size: 15.5px;
  line-height: 1.5;
  color: var(--prose-meta);
}

.blog-prose p > img {
  max-width: 100%;
  height: auto;
  border-radius: 12px;
}

.blog-prose hr {
  margin: 56px 0 0;
  border: 0;
  border-top: 1px solid var(--prose-rule);
}

.blog-prose table {
  display: block;
  max-width: 100%;
  margin: 30px 0 0;
  overflow-x: auto;
  border-collapse: collapse;
  font-size: .85em;
  line-height: 1.5;
}

.blog-prose :is(th, td) {
  padding: 12px 18px 12px 0;
  border-top: 1px solid var(--prose-rule);
  text-align: left;
  vertical-align: top;
}

.blog-prose thead th {
  border-top: 0;
  font-weight: 650;
  color: var(--prose-ink);
}

.blog-prose-closing {
  font-size: 17px;
  line-height: 1.6;
}

.blog-prose.blog-prose-closing p {
  margin: 0 0 12px;
}

@media (min-width: 640px) {
  .blog-prose {
    font-size: 20px;
  }

  .blog-prose p {
    margin-top: 26px;
  }

  .blog-prose h2 {
    margin-top: 80px;
  }

  .blog-prose blockquote {
    margin-top: 44px;
    padding: 34px 0 30px;
  }

  .blog-prose figure {
    margin-top: 48px;
  }

  .blog-prose figure img {
    padding: 28px 24px;
    border-radius: 24px;
  }

  .blog-prose figcaption {
    margin-top: 14px;
  }

  .blog-prose-closing {
    font-size: 18px;
  }
}

@media (min-width: 1300px) {
  .blog-prose figure {
    margin-right: -168px;
  }
}
</style>
