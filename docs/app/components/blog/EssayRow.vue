<script setup lang="ts">
import { formatDate, type PostSummary } from './post';

defineProps<{
  post: PostSummary;
  arrow?: boolean;
}>();
</script>

<template>
  <NuxtLink :to="post.path" class="row group border-t border-stone-300 dark:border-stone-800" :class="{ 'row-arrow': arrow }">
    <time :datetime="post.date" class="when text-[15.5px] whitespace-nowrap text-stone-600 dark:text-stone-400">{{ formatDate(post.date) }}</time>
    <div class="min-w-0">
      <h3 class="font-display text-3xl/tight text-balance text-stone-900 decoration-[1.5px] underline-offset-[5px] group-hover:underline lg:text-4xl/tight dark:text-stone-100">{{ post.title }}</h3>
      <p v-if="post.description" class="mt-2.5 max-w-[560px] text-[17px]/[1.55] text-stone-700 dark:text-stone-400">{{ post.description }}</p>
      <span class="mt-3 block text-[15px] text-stone-600 dark:text-stone-400">
        <template v-if="post.author">{{ post.author }} · </template>{{ post.minutes }} min read
      </span>
    </div>
    <span v-if="arrow" class="trail flex min-h-11 items-center text-stone-900 dark:text-stone-100" aria-hidden="true">
      <UIcon name="i-lucide-arrow-right" class="size-5" />
    </span>
    <div v-else class="trail thumb" :class="{ sketch: !post.image }" aria-hidden="true">
      <img v-if="post.image" :src="post.image.src" alt="" loading="lazy" decoding="async">
      <img v-else src="/img/landing/sketch-eject.svg" alt="" width="380" height="150" loading="lazy" decoding="async">
    </div>
  </NuxtLink>
</template>

<style scoped>
.row {
  display: grid;
  grid-template-columns: 132px minmax(0, 1fr) 150px;
  gap: 8px 28px;
  align-items: start;
  padding: 30px 0;
}

.row-arrow {
  grid-template-columns: 132px minmax(0, 1fr) auto;
}

.when {
  line-height: 44px;
}

.thumb {
  width: 150px;
  height: 126px;
  overflow: hidden;
  border-radius: 14px;
  background: #f6ede0;
  box-shadow: inset 0 0 0 1px oklch(0.25 0.04 55 / .08);
}

.thumb img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.thumb.sketch {
  background: oklch(0.928 0.026 8);
}

.thumb.sketch img {
  width: auto;
  max-width: none;
  height: 150px;
  margin: -10px 0 0 -212px;
}

@media (max-width: 760px) {
  .row {
    grid-template-columns: minmax(0, 1fr);
    padding: 26px 0;
  }

  .when {
    line-height: 1.4;
  }

  .trail {
    display: none;
  }
}
</style>
