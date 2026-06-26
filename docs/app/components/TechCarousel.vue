<script setup lang="ts">
const icons = [
  'devicon-nodejs-plain-wordmark',
  'devicon-typescript-plain',
  'devicon-bun-plain',
  'devicon-vitejs-plain',
  'devicon-nextjs-plain-wordmark',
  'devicon-nuxtjs-plain-wordmark',
  'devicon-astro-plain-wordmark',
  'devicon-angular-plain-wordmark',
  'devicon-react-plain-wordmark',
  'devicon-vuejs-plain-wordmark',
  'devicon-svelte-plain-wordmark',
  'devicon-remix-plain-wordmark',
  'devicon-express-original-wordmark',
  'devicon-python-plain-wordmark',
  'devicon-django-plain-wordmark',
  'devicon-flask-plain-wordmark',
  'devicon-fastapi-plain-wordmark',
  'devicon-php-plain',
  'devicon-laravel-plain-wordmark',
  'devicon-ruby-plain-wordmark',
  'devicon-rails-plain-wordmark',
  'devicon-java-plain-wordmark',
  'devicon-spring-plain-wordmark',
  'devicon-elixir-plain-wordmark',
  'devicon-phoenix-plain-wordmark',
  'devicon-go-plain-wordmark',
  'devicon-rust-plain',
  'devicon-denojs-plain-wordmark',
  'devicon-dotnetcore-plain',
  'devicon-gleam-plain',
  'devicon-cplusplus-plain-wordmark',
];

const loop = [...icons, ...icons];
</script>

<template>
  <section>
    <p class="px-6 text-center text-lg font-medium text-neutral-600 mb-12 dark:text-neutral-400">
      Lucity supports your favorite technology
    </p>

    <div class="group overflow-hidden mask-[linear-gradient(to_right,transparent,#000_6%,#000_94%,transparent)]">
      <div class="flex gap-20 w-max items-center animate-tech-scroll group-hover:[animation-play-state:paused] motion-reduce:w-full motion-reduce:animate-none motion-reduce:flex-wrap motion-reduce:justify-center motion-reduce:gap-y-10">
        <i
          v-for="(cls, i) in loop"
          :key="i"
          :class="cls"
          class="shrink-0 max-w-12 w-full text-7xl text-neutral-800 dark:text-neutral-200"
          aria-hidden="true"
        />
      </div>
    </div>
  </section>
</template>
