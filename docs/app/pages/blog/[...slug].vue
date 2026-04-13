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
  font-family: var(--font-serif);
  font-size: 2.25rem;
  font-weight: normal;
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
