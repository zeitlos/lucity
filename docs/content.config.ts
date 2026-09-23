import { fileURLToPath } from 'node:url';
import { defineContentConfig, defineCollection, z } from '@nuxt/content';

export default defineContentConfig({
  collections: {
    // Product documentation: technical, versioned with the platform.
    docs: defineCollection({
      type: 'page',
      source: {
        include: '**',
        exclude: ['index.md', 'blog/**'],
        prefix: '/',
      },
      schema: z.object({
        links: z.array(z.object({
          label: z.string(),
          icon: z.string(),
          to: z.string(),
          target: z.string().optional(),
        })).optional(),
      }),
    }),

    // Audience-facing pages: use cases, comparisons, and anything else that
    // sells rather than instructs. Lives outside content/ so the docs stay
    // technical, but keeps its URLs (/use-cases/*, /comparisons/*).
    solutions: defineCollection({
      type: 'page',
      source: {
        cwd: fileURLToPath(new URL('./solutions', import.meta.url)),
        include: '**',
        exclude: ['**/.navigation.yml'],
        prefix: '/',
      },
      schema: z.object({
        actions: z.array(z.object({
          label: z.string(),
          to: z.string(),
        })).optional(),
      }),
    }),

    blog: defineCollection({
      type: 'page',
      source: {
        include: 'blog/**',
        exclude: ['blog/.navigation.yml'],
        prefix: '/blog',
      },
      schema: z.object({
        date: z.string(),
        author: z.string().optional(),
      }),
    }),
  },
});
