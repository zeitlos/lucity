<script setup lang="ts">
interface Guide {
  title: string;
  to: string;
  runtime: string;
  blurb: string;
  icon: string;
  color: string;
  colorDark?: string;
  iconClass?: string;
}

const guides: Guide[] = [
  {
    title: 'Next.js',
    to: '/guides/nextjs',
    runtime: 'Node',
    blurb: 'App Router, server actions, and a PostgreSQL database.',
    icon: 'i-devicon-nextjs',
    color: '#111111',
    colorDark: '#D4D4D4',
  },
  {
    title: 'Nuxt',
    to: '/guides/nuxt',
    runtime: 'Node',
    blurb: 'Nitro output, runtime config, and server routes.',
    icon: 'i-devicon-nuxtjs',
    color: '#00DC82',
  },
  {
    title: 'SvelteKit',
    to: '/guides/sveltekit',
    runtime: 'Node',
    blurb: 'The Node adapter, and what it takes to deploy.',
    icon: 'i-devicon-svelte',
    color: '#FF3E00',
  },
  {
    title: 'Astro',
    to: '/guides/astro',
    runtime: 'Node',
    blurb: 'Static output or server rendering, your choice.',
    icon: 'i-devicon-astro',
    color: '#BC52EE',
  },
  {
    title: 'Express',
    to: '/guides/express',
    runtime: 'Node',
    blurb: 'A plain HTTP API, with health checks that pass.',
    icon: 'i-simple-icons-express',
    color: '#4B5563',
    colorDark: '#C4C4C4',
    iconClass: 'text-neutral-800',
  },
  {
    title: 'FastAPI',
    to: '/guides/fastapi',
    runtime: 'Python',
    blurb: 'Uvicorn, async database access, and migrations.',
    icon: 'i-devicon-fastapi',
    color: '#009688',
  },
  {
    title: 'Django',
    to: '/guides/django',
    runtime: 'Python',
    blurb: 'Gunicorn, migrations, and static files in a bucket.',
    icon: 'i-simple-icons-django',
    color: '#0C4B33',
    colorDark: '#44B78B',
    iconClass: 'text-[#0C4B33]',
  },
  {
    title: 'Laravel',
    to: '/guides/laravel',
    runtime: 'PHP',
    blurb: 'Artisan on release, plus a queue worker alongside.',
    icon: 'i-devicon-laravel',
    color: '#FF2D20',
  },
  {
    title: 'Go',
    to: '/guides/go',
    runtime: 'Go',
    blurb: 'One module in, one small binary out.',
    icon: 'i-devicon-go',
    color: '#00ADD8',
  },
];
</script>

<template>
  <div class="guide-cards">
    <NuxtLink
      v-for="guide in guides"
      :key="guide.to"
      :to="guide.to"
      class="guide-card group"
      :style="{ '--brand': guide.color, '--brand-dark': guide.colorDark || guide.color }"
    >
      <span class="guide-card-glow" />

      <span class="guide-card-chip">
        <UIcon
          :name="guide.icon"
          class="size-7"
          :class="guide.iconClass"
        />
      </span>

      <UBadge
        :label="guide.runtime"
        color="neutral"
        variant="outline"
        size="sm"
        class="guide-card-badge"
      />

      <span class="guide-card-title">{{ guide.title }}</span>

      <span class="guide-card-blurb">{{ guide.blurb }}</span>

      <span class="guide-card-cta">
        Read the guide
        <UIcon
          name="i-lucide-arrow-right"
          class="size-4 transition-transform group-hover:translate-x-0.5"
        />
      </span>
    </NuxtLink>
  </div>
</template>

<style scoped>
.guide-cards {
  display: grid;
  gap: 1rem;
  margin: 2rem 0;
  grid-template-columns: repeat(1, minmax(0, 1fr));
}

@media (min-width: 640px) {
  .guide-cards {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1280px) {
  .guide-cards {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.guide-card {
  --accent: var(--brand);
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1.25rem;
  border-radius: 0.875rem;
  border: 1px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  box-shadow: 0 1px 2px oklch(0 0 0 / 0.04);
  text-decoration: none;
  transition: transform 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.guide-card:hover {
  transform: translateY(-2px);
  border-color: color-mix(in oklab, var(--accent) 45%, transparent);
  box-shadow: 0 6px 18px -14px color-mix(in oklab, var(--accent) 55%, transparent);
}

.guide-card-glow {
  position: absolute;
  inset: -40% 30% 55% -20%;
  background: radial-gradient(60% 60% at 30% 70%, color-mix(in oklab, var(--accent) 42%, transparent) 0%, transparent 70%);
  filter: blur(28px);
  opacity: 0.75;
  transition: opacity 0.2s ease;
  pointer-events: none;
}

.guide-card:hover .guide-card-glow {
  opacity: 1;
}

.guide-card-chip {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 3rem;
  height: 3rem;
  border-radius: 0.75rem;
  background: #fff;
  border: 1px solid color-mix(in oklab, var(--accent) 25%, transparent);
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.06);
}

.guide-card-badge {
  position: absolute;
  top: 1rem;
  inset-inline-end: 1rem;
  z-index: 1;
}

.guide-card-title {
  position: relative;
  margin-top: 0.5rem;
  font-weight: 600;
  font-size: 1.0625rem;
  color: var(--ui-text-highlighted);
}

.guide-card-blurb {
  position: relative;
  font-size: 0.875rem;
  line-height: 1.5;
  color: var(--ui-text-muted);
  text-wrap: pretty;
}

.guide-card-cta {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  margin-top: auto;
  padding-top: 1rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: color-mix(in oklab, var(--accent) 78%, var(--ui-text-highlighted));
}
</style>

<style>
/* Unscoped: the scoped compiler drops the descendant part of :global(.dark). */
.dark .guide-card {
  --accent: var(--brand-dark, var(--brand));
}
</style>
