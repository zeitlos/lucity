import { defineContentConfig, defineCollection, z } from '@nuxt/content';

export default defineContentConfig({
  collections: {
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
