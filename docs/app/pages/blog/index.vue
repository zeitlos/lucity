<script setup lang="ts">
definePageMeta({
  layout: 'default',
});

useSeo({
  title: 'Blog',
  description: 'Updates, deep dives, and stories from the Lucity team.',
  type: 'website',
});

const { data: posts } = await useAsyncData('blog-index', () =>
  queryCollection('blog').order('date', 'DESC').all()
);

function formatDate(date: string) {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}
</script>

<template>
  <div class="blog-index">
    <UContainer>
      <div class="blog-header">
        <h1>Blog</h1>
        <p class="blog-subtitle">
          Updates, deep dives, and stories from building an ejectable PaaS.
        </p>
      </div>

      <div class="blog-posts">
        <NuxtLink
          v-for="post in posts"
          :key="post.path"
          :to="post.path"
          class="blog-post-entry"
        >
          <time :datetime="post.date" class="blog-post-date">
            {{ formatDate(post.date) }}
          </time>
          <h2 class="blog-post-title">
            {{ post.title }}
          </h2>
          <p
            v-if="post.description"
            class="blog-post-description"
          >
            {{ post.description }}
          </p>
        </NuxtLink>
      </div>
    </UContainer>
  </div>
</template>

<style scoped>
.blog-index {
  padding-top: 4rem;
  padding-bottom: 6rem;
}

.blog-header {
  max-width: 680px;
  margin: 0 auto 4rem;
  text-align: center;
}

.blog-header h1 {
  font-family: var(--font-serif);
  font-size: 3.5rem;
  font-weight: normal;
  color: var(--ui-text-highlighted);
  margin-bottom: 1rem;
}

.blog-subtitle {
  font-size: 1.125rem;
  color: var(--ui-text-muted);
}

.blog-posts {
  max-width: 680px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.blog-post-entry {
  display: block;
  padding: 2rem 0;
  border-top: 1px solid var(--ui-border);
  text-decoration: none;
  transition: opacity 0.15s;
}

.blog-post-entry:last-child {
  border-bottom: 1px solid var(--ui-border);
}

.blog-post-entry:hover {
  opacity: 0.75;
}

.blog-post-date {
  display: block;
  font-size: 0.875rem;
  color: var(--ui-text-muted);
  margin-bottom: 0.5rem;
}

.blog-post-title {
  font-family: var(--font-serif);
  font-size: 1.75rem;
  font-weight: normal;
  color: var(--ui-text-highlighted);
  margin: 0;
  line-height: 1.3;
}

.blog-post-description {
  font-size: 1rem;
  color: var(--ui-text-muted);
  margin-top: 0.5rem;
  line-height: 1.6;
}
</style>
