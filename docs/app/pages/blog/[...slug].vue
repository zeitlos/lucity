<script setup lang="ts">
import { kebabCase } from 'scule';

definePageMeta({
  layout: 'default',
});

const route = useRoute();

const { data: page } = await useAsyncData(kebabCase(route.path), () =>
  queryCollection('blog').path(route.path).first()
);

if (!page.value) {
  throw createError({ statusCode: 404, statusMessage: 'Post not found', fatal: true });
}

function walk(node: unknown, counts: { words: number; images: number }): void {
  if (node == null) return;
  if (typeof node === 'string') {
    counts.words += node.trim().split(/\s+/).filter(Boolean).length;
    return;
  }
  if (Array.isArray(node)) {
    // Minimark compact format: [tag, props, ...children] — skip tag and props
    if (typeof node[0] === 'string' && (node[1] === null || typeof node[1] === 'object')) {
      if (node[0] === 'img') counts.images += 1;
      for (let i = 2; i < node.length; i++) walk(node[i], counts);
      return;
    }
    for (const child of node) walk(child, counts);
    return;
  }
  if (typeof node === 'object') {
    const n = node as { type?: string; tag?: string; value?: unknown; children?: unknown };
    if (n.type === 'text' && typeof n.value === 'string') {
      counts.words += n.value.trim().split(/\s+/).filter(Boolean).length;
      return;
    }
    if (n.tag === 'img') counts.images += 1;
    if (n.value !== undefined) walk(n.value, counts);
    if (n.children !== undefined) walk(n.children, counts);
  }
}

const readingTime = computed(() => {
  const counts = { words: 0, images: 0 };
  walk(page.value?.body, counts);
  // Medium-style: 225 WPM, plus decrementing image time (12s first, down to 3s)
  const textSeconds = (counts.words / 225) * 60;
  let imageSeconds = 0;
  for (let i = 0; i < counts.images; i++) {
    imageSeconds += Math.max(12 - i, 3);
  }
  return Math.max(1, Math.ceil((textSeconds + imageSeconds) / 60));
});

const formattedDate = computed(() => {
  if (!page.value?.date) return null;
  return new Date(page.value.date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
});

useSeo({
  title: page.value.title,
  description: page.value.description,
  type: 'article',
});
</script>

<template>
  <div
    v-if="page"
    class="blog-article"
  >
    <UContainer>
      <article class="blog-article-inner">
        <NuxtLink
          to="/blog"
          class="blog-back"
        >
          <UIcon
            name="i-lucide-arrow-left"
            class="size-4"
          />
          Back to blog
        </NuxtLink>

        <header class="blog-article-header">
          <div class="blog-meta">
            <time
              v-if="formattedDate"
              :datetime="page.date"
            >
              {{ formattedDate }}
            </time>
            <template v-if="formattedDate && page.author">
              <span>&middot;</span>
            </template>
            <span v-if="page.author">{{ page.author }}</span>
            <span>&middot;</span>
            <span>{{ readingTime }} min read</span>
          </div>

          <h1>{{ page.title }}</h1>

          <p
            v-if="page.description"
            class="blog-lede"
          >
            {{ page.description }}
          </p>
        </header>

        <div class="blog-body prose prose-stone dark:prose-invert">
          <ContentRenderer :value="page" />
        </div>

        <footer class="blog-footer">
          <USeparator />
          <NuxtLink
            to="/blog"
            class="blog-back"
          >
            <UIcon
              name="i-lucide-arrow-left"
              class="size-4"
            />
            All posts
          </NuxtLink>
        </footer>
      </article>
    </UContainer>
  </div>
</template>

<style scoped>
.blog-article {
  padding-top: 3rem;
  padding-bottom: 6rem;
}

.blog-article-inner {
  max-width: 680px;
  margin: 0 auto;
}

.blog-back {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.875rem;
  color: var(--ui-text-muted);
  text-decoration: none;
  transition: color 0.15s;
}

.blog-back:hover {
  color: var(--ui-text-highlighted);
}

.blog-article-header {
  margin-top: 2rem;
  margin-bottom: 3rem;
}

.blog-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: var(--ui-text-muted);
  margin-bottom: 1rem;
}

.blog-article-header h1 {
  font-family: var(--font-serif);
  font-size: 2.25rem;
  font-weight: normal;
  color: var(--ui-text-highlighted);
  line-height: 1.2;
  margin: 0;
}

@media (min-width: 640px) {
  .blog-article-header h1 {
    font-size: 3rem;
  }
}

@media (min-width: 1024px) {
  .blog-article-header h1 {
    font-size: 3.75rem;
  }
}

.blog-lede {
  font-size: 1.25rem;
  color: var(--ui-text-muted);
  line-height: 1.6;
  margin-top: 1rem;
}

.blog-body {
  font-size: 1.0625rem;
  line-height: 1.8;
}

.blog-body :deep(h2) {
  font-family: var(--font-sans);
  font-size: 1.875rem;
  font-weight: 600;
  margin-top: 2.5rem;
}

.blog-body :deep(h3) {
  font-size: 1.25rem;
  margin-top: 2rem;
}

.blog-body :deep(p) {
  margin-top: 1.25rem;
  margin-bottom: 0;
}

.blog-body :deep(ul),
.blog-body :deep(ol) {
  margin-top: 1.25rem;
}

.blog-body :deep(img) {
  display: block;
  max-width: 100%;
  height: auto;
  margin: 2rem auto 0;
  border: 1px solid var(--ui-border);
  border-radius: 0.5rem;
  box-shadow: 0 4px 12px oklch(0 0 0 / 0.08);
}

.blog-body :deep(img + em) {
  display: block;
  text-align: center;
  font-style: normal;
  font-size: 0.875rem;
  color: var(--ui-text-muted);
  margin: 0.75rem 0 2rem;
}

.blog-footer {
  margin-top: 4rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
</style>

<style>
.blog-body p:first-of-type::first-letter {
  font-family: var(--font-serif);
  font-size: 3.75rem;
  float: left;
  line-height: 0.8;
  margin-right: 0.1em;
  margin-top: 0.1em;
  color: var(--ui-text-highlighted);
}
</style>
